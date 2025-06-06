package main

import (
	"archive/zip"
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"path"
	"strconv"

	"github.com/aws/smithy-go"
	"github.com/ybizeul/apiws/auth"
	"github.com/ybizeul/hupload/internal/storage"
)

// type ShareParameters struct {
// 	Exposure    string `json:"exposure"`
// 	Validity    int    `json:"validity"`
// 	Description string `json:"description"`
// 	Message     string `json:"message"`
// }

// postShare creates a new share with a randomly generate name
func (h *Hupload) postShare(w http.ResponseWriter, r *http.Request) {
	user, _ := auth.UserForRequest(r)

	// This should never happen as authentication is checked before in the
	// middleware
	if user == "" {
		slog.Error("postShare", slog.String("error", "no user in context"))
		http.Error(w, "no user in context", http.StatusBadRequest)
		return
	}

	code := r.PathValue("share")
	if code == "" {
		code = generateCode(4, 3)
	}

	// Parse the request body
	options := storage.Options{
		Exposure: "upload",
		Validity: h.Config.Values.DefaultValidityDays,
	}

	// We ignore unmarshalling of JSON body as it is optional.
	_ = json.NewDecoder(r.Body).Decode(&options)

	share, err := h.Config.Storage.CreateShare(r.Context(), code, user, options)
	if err != nil {
		slog.Error("postShare", slog.String("error", err.Error()))
		switch {
		case errors.Is(err, storage.ErrShareAlreadyExists):
			writeError(w, http.StatusConflict, "share already exists")
			return
		}

		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeSuccessJSON(w, share)
}

// patchShare updates an existing share
func (h *Hupload) patchShare(w http.ResponseWriter, r *http.Request) {
	user, _ := auth.UserForRequest(r)

	// This should never happen as authentication is checked before in the
	// middleware
	if user == "" {
		slog.Error("patchShare", slog.String("error", "no user in context"))
		http.Error(w, "no user in context", http.StatusBadRequest)
		return
	}

	// Parse the request body
	options := &storage.Options{}

	// We ignore unmarshalling of JSON body as it is optional.
	_ = json.NewDecoder(r.Body).Decode(&options)

	share, err := h.Config.Storage.GetShare(r.Context(), r.PathValue("share"))
	if err != nil {
		slog.Error("patchShare", slog.String("error", err.Error()))
		switch {
		case errors.Is(err, storage.ErrShareNotFound):
			writeError(w, http.StatusNotFound, "share not found")
			return
		}
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	if h.Config.Values.HideOtherShares {
		if share.Owner != user {
			writeError(w, http.StatusForbidden, "unauthorized")
			return
		}
	}

	result, err := h.Config.Storage.UpdateShare(r.Context(), share.Name, options)
	if err != nil {
		slog.Error("patchShare", slog.String("error", err.Error()))
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeSuccessJSON(w, result)
}

// postItem copies a new item in the share and returns the json description
func (h *Hupload) postItem(w http.ResponseWriter, r *http.Request) {
	share, err := h.Config.Storage.GetShare(r.Context(), r.PathValue("share"))
	if err != nil {
		slog.Error("postItem", slog.String("error", err.Error()))
		switch {
		case errors.Is(err, storage.ErrShareNotFound):
			writeError(w, http.StatusNotFound, "share not found")
			return
		}
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	user, _ := auth.UserForRequest(r)

	if user == "" && (share.Options.Exposure != "both" && share.Options.Exposure != "upload") {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	mp, err := r.MultipartReader()
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	np, err := mp.NextPart()
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	cl := 0

	if r.Header.Get("FileSize") == "" {
		writeError(w, http.StatusBadRequest, "missing content length")
	}

	cl, err = strconv.Atoi(r.Header.Get("FileSize"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid content length")
		return
	}

	b := bufio.NewReader(np)
	item, err := h.Config.Storage.CreateItem(r.Context(), r.PathValue("share"), r.PathValue("item"), int64(cl), b)
	var apiErr smithy.APIError
	if err != nil {
		switch {
		case errors.Is(err, storage.ErrMaxShareSizeReached):
			writeError(w, http.StatusInsufficientStorage, "max share size reached")
			return
		case errors.Is(err, storage.ErrMaxFileSizeReached):
			writeError(w, http.StatusInsufficientStorage, "max item size reached")
			return
		}
		if errors.As(err, &apiErr) {
			writeError(w, http.StatusBadRequest, apiErr.ErrorMessage())
			return
		}
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeSuccessJSON(w, item)
}

// postItem copies a new item in the share and returns the json description
func (h *Hupload) deleteItem(w http.ResponseWriter, r *http.Request) {
	share, err := h.Config.Storage.GetShare(r.Context(), r.PathValue("share"))
	if err != nil {
		slog.Error("postItem", slog.String("error", err.Error()))
		switch {
		case errors.Is(err, storage.ErrShareNotFound):
			writeError(w, http.StatusNotFound, "share not found")
			return
		}
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	user, _ := auth.UserForRequest(r)

	if user == "" && (share.Options.Exposure != "both" && share.Options.Exposure != "upload") {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	err = h.Config.Storage.DeleteItem(r.Context(), r.PathValue("share"), r.PathValue("item"))
	if err != nil {
		switch {
		case errors.Is(err, storage.ErrItemNotFound):
			writeError(w, http.StatusNotFound, "item does not exists")
			return
		}
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeSuccess(w, "item deleted")
}

// getShares returns the list of shares as json
func (h *Hupload) getShares(w http.ResponseWriter, r *http.Request) {
	shares, err := h.Config.Storage.ListShares(r.Context())

	if err != nil {
		slog.Error("getShares", slog.String("error", err.Error()))
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	user, _ := auth.UserForRequest(r)

	if h.Config.Values.HideOtherShares {
		tmpShares := []storage.Share{}
		for s := range shares {
			if shares[s].Owner == user {
				tmpShares = append(tmpShares, shares[s])
			}
		}
		shares = tmpShares
	}

	if user == "" {
		writeSuccessJSON(w, storage.PublicShares(shares))
	} else {
		writeSuccessJSON(w, shares)
	}
}

// getShareItems returns the share identified by the request parameter
func (h *Hupload) getShare(w http.ResponseWriter, r *http.Request) {
	share, err := h.Config.Storage.GetShare(r.Context(), r.PathValue("share"))

	if err != nil {
		slog.Error("getShare", slog.String("error", err.Error()))
		switch {
		case errors.Is(err, storage.ErrShareNotFound):
			writeError(w, http.StatusNotFound, "share not found")
			return
		}
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	user, _ := auth.UserForRequest(r)

	if user == "" && !share.IsValid() {
		writeError(w, http.StatusGone, "Share expired")
		return
	}

	if user == "" {
		writeSuccessJSON(w, share.PublicShare())
	} else {
		writeSuccessJSON(w, share)
	}
}

// getShareItems returns the share content identified by the request parameter
func (h *Hupload) getShareItems(w http.ResponseWriter, r *http.Request) {
	share, err := h.Config.Storage.GetShare(r.Context(), r.PathValue("share"))
	if err != nil {
		slog.Error("getShareItems", slog.String("error", err.Error()))
		switch {
		case errors.Is(err, storage.ErrShareNotFound):
			writeError(w, http.StatusNotFound, "share not found")
			return
		}
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	user, _ := auth.UserForRequest(r)

	if user == "" && !share.IsValid() {
		writeError(w, http.StatusGone, "Share expired")
		return
	}

	content, err := h.Config.Storage.ListShare(r.Context(), share.Name)
	if err != nil {
		slog.Error("getShareItems", slog.String("error", err.Error()))
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeSuccessJSON(w, content)
}

// deleteShare deletes the share identified by the request parameter
func (h *Hupload) deleteShare(w http.ResponseWriter, r *http.Request) {
	err := h.Config.Storage.DeleteShare(r.Context(), r.PathValue("share"))
	if err != nil {
		slog.Error("deleteShare", slog.String("error", err.Error()))
		switch {
		case errors.Is(err, storage.ErrShareNotFound):
			writeError(w, http.StatusNotFound, "share not found")
			return
		}
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeSuccess(w, "share deleted")
}

// getItem returns the item identified by the request parameter
func (h *Hupload) getItem(w http.ResponseWriter, r *http.Request) {
	shareName := r.PathValue("share")
	itemName := r.PathValue("item")

	share, err := h.Config.Storage.GetShare(r.Context(), shareName)
	if err != nil {
		slog.Error("getItem", slog.String("error", err.Error()))
		switch {
		case errors.Is(err, storage.ErrShareNotFound):
			writeError(w, http.StatusNotFound, "share not found")
			return
		}
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	user, _ := auth.UserForRequest(r)

	if user == "" && (share.Options.Exposure != "both" && share.Options.Exposure != "download") {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	item, err := h.Config.Storage.GetItem(r.Context(), shareName, itemName)
	if err != nil {
		switch {
		case errors.Is(err, storage.ErrItemNotFound):
			writeError(w, http.StatusNotFound, err.Error())
			return
		}
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	reader, err := h.Config.Storage.GetItemData(r.Context(), shareName, itemName)
	if err != nil {
		slog.Error("getShares", slog.String("error", err.Error()))
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	defer reader.Close()

	w.Header().Add("Content-Type", "application/octet-stream")
	w.Header().Add("Content-Length", fmt.Sprintf("%d", item.ItemInfo.Size))
	w.Header().Add("Content-Disposition", "attachment")

	_, err = io.Copy(w, reader)
	if err != nil {
		slog.Error("getItem", slog.String("error", err.Error()))
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
}

// getVersion returns hupload version
func (h *Hupload) downloadShare(w http.ResponseWriter, r *http.Request) {
	shareName := r.PathValue("share")

	share, err := h.Config.Storage.GetShare(r.Context(), shareName)
	if err != nil {
		slog.Error("getItem", slog.String("error", err.Error()))
		switch {
		case errors.Is(err, storage.ErrShareNotFound):
			writeError(w, http.StatusNotFound, "share not found")
			return
		}
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	user, _ := auth.UserForRequest(r)

	if user == "" && (share.Options.Exposure != "both" && share.Options.Exposure != "download") {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	items, err := h.Config.Storage.ListShare(r.Context(), shareName)
	if err != nil {
		slog.Error("downloadShare", slog.String("error", err.Error()))
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	w.Header().Add("Content-Type", "application/zip")
	w.Header().Add("Content-Disposition", "attachment")

	zipWriter := zip.NewWriter(w)

	for _, item := range items {
		f, err := zipWriter.Create(path.Base(item.Path))
		if err != nil {
			slog.Error("downloadShare", slog.String("error", err.Error()))
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		d, err := h.Config.Storage.GetItemData(r.Context(), shareName, path.Base(item.Path))
		if err != nil {
			slog.Error("downloadShare", slog.String("error", err.Error()))
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		defer d.Close()
		_, err = io.Copy(f, d)
		if err != nil {
			slog.Error("downloadShare", slog.String("error", err.Error()))
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
	}
	err = zipWriter.Close()
	if err != nil {
		slog.Error("downloadShare", slog.String("error", err.Error()))
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
}

// postLogin returns the user name for the current session
func (h *Hupload) postLogin(w http.ResponseWriter, r *http.Request) {
	user, _ := auth.UserForRequest(r)
	//	_, loginURL := h.Config.Authentication.LoginURL(nil)
	u := struct {
		User string `json:"user"`
		//		LoginPage string `json:"loginPage"`
	}{
		User: user,
		//		LoginPage: loginURL,
	}
	writeSuccessJSON(w, u)
}

func (h *Hupload) getMessages(w http.ResponseWriter, r *http.Request) {
	titles := []string{}

	for _, m := range h.Config.Values.MessageTemplates {
		titles = append(titles, m.Title)
	}

	writeSuccessJSON(w, titles)
}

func (h *Hupload) getDefaults(w http.ResponseWriter, r *http.Request) {
	defaults := struct {
		Validity int    `json:"validity"`
		Exposure string `json:"exposure"`
	}{
		Validity: h.Config.Values.DefaultValidityDays,
		Exposure: h.Config.Values.DefaultExposure,
	}

	writeSuccessJSON(w, defaults)
}

var ErrMessageInvalidIndex = errors.New("invalid index")
var ErrMessageIndexOutOfBounds = errors.New("index out of bounds")

func (h *Hupload) getMessage(w http.ResponseWriter, r *http.Request) {
	index, err := strconv.Atoi(r.PathValue("index"))
	if err != nil {
		writeError(w, http.StatusBadRequest, ErrMessageInvalidIndex.Error())
		return
	}
	t := h.Config.Values.MessageTemplates
	if index <= len(t) && index > 0 {
		writeSuccessJSON(w, t[index-1])
		return
	} else {
		writeError(w, http.StatusBadRequest, ErrMessageIndexOutOfBounds.Error())
		return
	}
}

// getVersion returns hupload version
func (h *Hupload) getVersion(w http.ResponseWriter, r *http.Request) {
	v := struct {
		Version string `json:"version"`
	}{
		Version: version,
	}
	writeSuccessJSON(w, v)
}
