package main

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"

	"github.com/ybizeul/hupload/internal/storage"
	"github.com/ybizeul/hupload/pkg/apiws/middleware/auth"
)

// type ShareParameters struct {
// 	Exposure    string `json:"exposure"`
// 	Validity    int    `json:"validity"`
// 	Description string `json:"description"`
// 	Message     string `json:"message"`
// }

// postShare creates a new share with a randomly generate name
func postShare(w http.ResponseWriter, r *http.Request) {
	user := auth.UserForRequest(r)
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
		Validity: cfg.Values.DefaultValidityDays,
	}

	// We ignore unmarshalling of JSON body as it is optional.
	_ = json.NewDecoder(r.Body).Decode(&options)

	share, err := cfg.Storage.CreateShare(code, user, options)
	if err != nil {
		slog.Error("postShare", slog.String("error", err.Error()))
		switch {
		case errors.Is(err, storage.ErrInvalidShareName):
			writeError(w, http.StatusBadRequest, "invalid share name")
			return
		case errors.Is(err, storage.ErrShareAlreadyExists):
			writeError(w, http.StatusConflict, "share already exists")
			return
		}

		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeSuccessJSON(w, share)
}

// putShare creates a new share with name from the request parameter
// func putShare(w http.ResponseWriter, r *http.Request) {
// 	user := auth.UserForRequest(r)
// 	if user == "" {
// 		slog.Error("putShare", slog.String("error", "no user in context"))
// 		writeError(w, http.StatusUnauthorized, "no user in context")
// 		return
// 	}

// 	// Parse the request body
// 	params := ShareParameters{
// 		Exposure: "upload",
// 		Validity: cfg.Values.DefaultValidityDays,
// 	}

// 	_ = json.NewDecoder(r.Body).Decode(&params)

// 	share, err := cfg.Storage.CreateShare(r.PathValue("share"), user, storage.Options{Validity: params.Validity, Exposure: params.Exposure})
// 	if err != nil {
// 		slog.Error("putShare", slog.String("error", err.Error()))
// 		switch {
// 		case errors.Is(err, storage.ErrInvalidShareName):
// 			writeError(w, http.StatusBadRequest, "invalid share name")
// 			return
// 		case errors.Is(err, storage.ErrShareNotFound):
// 			writeError(w, http.StatusNotFound, "share not found")
// 			return
// 		case errors.Is(err, storage.ErrShareAlreadyExists):
// 			writeError(w, http.StatusConflict, "share already exists")
// 			return
// 		}
// 		writeError(w, http.StatusInternalServerError, err.Error())
// 		return
// 	}
// 	writeSuccessJSON(w, share)
// }

// postItem copies a new item in the share and returns the json description
func postItem(w http.ResponseWriter, r *http.Request) {
	share, err := cfg.Storage.GetShare(r.PathValue("share"))
	if err != nil {
		slog.Error("postItem", slog.String("error", err.Error()))
		switch {
		case errors.Is(err, storage.ErrInvalidShareName):
			writeError(w, http.StatusBadRequest, "invalid share name")
			return
		case errors.Is(err, storage.ErrShareNotFound):
			writeError(w, http.StatusNotFound, "share not found")
			return
		}
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	if auth.UserForRequest(r) == "" && (share.Options.Exposure != "both" && share.Options.Exposure != "upload") {
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
	b := bufio.NewReader(np)
	item, err := cfg.Storage.CreateItem(r.PathValue("share"), r.PathValue("item"), b)
	if err != nil {
		switch {
		case errors.Is(err, storage.ErrInvalidShareName):
			writeError(w, http.StatusBadRequest, "invalid share name")
			return
		case errors.Is(err, storage.ErrShareNotFound):
			writeError(w, http.StatusNotFound, "share not found")
			return
		case errors.Is(err, storage.ErrMaxShareSizeReached):
			writeError(w, http.StatusInsufficientStorage, "max share size reached")
			return
		}
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeSuccessJSON(w, item)
}

// postItem copies a new item in the share and returns the json description
func deleteItem(w http.ResponseWriter, r *http.Request) {
	share, err := cfg.Storage.GetShare(r.PathValue("share"))
	if err != nil {
		slog.Error("postItem", slog.String("error", err.Error()))
		switch {
		case errors.Is(err, storage.ErrInvalidShareName):
			writeError(w, http.StatusBadRequest, "invalid share name")
			return
		case errors.Is(err, storage.ErrShareNotFound):
			writeError(w, http.StatusNotFound, "share not found")
			return
		}
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	if auth.UserForRequest(r) == "" && (share.Options.Exposure != "both" && share.Options.Exposure != "upload") {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	err = cfg.Storage.DeleteItem(r.PathValue("share"), r.PathValue("item"))
	if err != nil {
		switch {
		case errors.Is(err, storage.ErrInvalidShareName):
			writeError(w, http.StatusBadRequest, "invalid share name")
			return
		case errors.Is(err, storage.ErrShareNotFound):
			writeError(w, http.StatusNotFound, "share not found")
			return
		case errors.Is(err, storage.ErrInvalidItemName):
			writeError(w, http.StatusBadRequest, "invalid item name")
			return
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
func getShares(w http.ResponseWriter, r *http.Request) {
	shares, err := cfg.Storage.ListShares()

	if err != nil {
		slog.Error("getShares", slog.String("error", err.Error()))
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeSuccessJSON(w, shares)
}

// getShareItems returns the share identified by the request parameter
func getShare(w http.ResponseWriter, r *http.Request) {
	share, err := cfg.Storage.GetShare(r.PathValue("share"))

	if err != nil {
		slog.Error("getShare", slog.String("error", err.Error()))
		switch {
		case errors.Is(err, storage.ErrInvalidShareName):
			writeError(w, http.StatusBadRequest, "invalid share name")
			return
		case errors.Is(err, storage.ErrShareNotFound):
			writeError(w, http.StatusNotFound, "share not found")
			return
		}
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	if auth.UserForRequest(r) == "" && !share.IsValid() {
		writeError(w, http.StatusGone, "Share expired")
		return
	}

	writeSuccessJSON(w, share)
}

// getShareItems returns the share content identified by the request parameter
func getShareItems(w http.ResponseWriter, r *http.Request) {
	share, err := cfg.Storage.GetShare(r.PathValue("share"))
	if err != nil {
		slog.Error("getShareItems", slog.String("error", err.Error()))
		switch {
		case errors.Is(err, storage.ErrInvalidShareName):
			writeError(w, http.StatusBadRequest, "invalid share name")
			return
		case errors.Is(err, storage.ErrShareNotFound):
			writeError(w, http.StatusNotFound, "share not found")
			return
		}
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	if auth.UserForRequest(r) == "" && !share.IsValid() {
		writeError(w, http.StatusGone, "Share expired")
		return
	}

	content, err := cfg.Storage.ListShare(share.Name)
	if err != nil {
		slog.Error("getShareItems", slog.String("error", err.Error()))
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeSuccessJSON(w, content)
}

// deleteShare deletes the share identified by the request parameter
func deleteShare(w http.ResponseWriter, r *http.Request) {
	err := cfg.Storage.DeleteShare(r.PathValue("share"))
	if err != nil {
		slog.Error("deleteShare", slog.String("error", err.Error()))
		switch {
		case errors.Is(err, storage.ErrInvalidShareName):
			writeError(w, http.StatusBadRequest, "invalid share name")
			return
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
func getItem(w http.ResponseWriter, r *http.Request) {
	shareName := r.PathValue("share")
	itemName := r.PathValue("item")

	share, err := cfg.Storage.GetShare(shareName)
	if err != nil {
		slog.Error("getItem", slog.String("error", err.Error()))
		switch {
		case errors.Is(err, storage.ErrInvalidShareName):
			writeError(w, http.StatusBadRequest, "invalid share name")
			return
		case errors.Is(err, storage.ErrShareNotFound):
			writeError(w, http.StatusNotFound, "share not found")
			return
		}
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	if auth.UserForRequest(r) == "" && (share.Options.Exposure != "both" && share.Options.Exposure != "download") {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	item, err := cfg.Storage.GetItem(shareName, itemName)
	if err != nil {
		switch {
		case errors.Is(err, storage.ErrItemNotFound):
			writeError(w, http.StatusNotFound, err.Error())
			return
		case errors.Is(err, storage.ErrInvalidItemName):
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	reader, err := cfg.Storage.GetItemData(shareName, itemName)
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

// postLogin returns the user name for the current session
func postLogin(w http.ResponseWriter, r *http.Request) {
	u := struct {
		User string `json:"user"`
	}{
		User: auth.UserForRequest(r),
	}
	writeSuccessJSON(w, u)
}

// getVersion returns hupload version
func getVersion(w http.ResponseWriter, r *http.Request) {
	v := struct {
		Version string `json:"version"`
	}{
		Version: version,
	}
	writeSuccessJSON(w, v)
}
