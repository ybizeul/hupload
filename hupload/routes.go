package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"

	"github.com/ybizeul/hupload/pkg/apiws/middleware/auth"
)

// postShare creates a new share with a randomly generate name
func postShare(w http.ResponseWriter, r *http.Request) {
	user := auth.UserForRequest(r)
	if user == "" {
		slog.Error("postShare", slog.String("error", "no user in context"))
		_, _ = w.Write([]byte("no user in context"))
		return
	}
	code := generateCode(4, 3)
	err := cfg.Storage.CreateShare(code, user, cfg.Values.DefaultValidityDays)
	if err != nil {
		slog.Error("postShare", slog.String("error", err.Error()))
		_, _ = w.Write([]byte(err.Error()))
		return
	}
	_, _ = w.Write([]byte(code))

}

// putShare creates a new share with name from the request parameter
func putShare(w http.ResponseWriter, r *http.Request) {
	user := auth.UserForRequest(r)
	if user == "" {
		slog.Error("putShare", slog.String("error", "no user in context"))
		_, _ = w.Write([]byte("no user in context"))
		return
	}
	err := cfg.Storage.CreateShare(r.PathValue("share"), user, cfg.Values.DefaultValidityDays)
	if err != nil {
		slog.Error("putShare", slog.String("error", err.Error()))
		_, _ = w.Write([]byte(err.Error()))
		return
	}
}

// postItem copies a new item in the share and returns the json description
func postItem(w http.ResponseWriter, r *http.Request) {
	mp, err := r.MultipartReader()
	if err != nil {
		_, _ = w.Write([]byte(err.Error()))
		return
	}
	np, err := mp.NextPart()
	if err != nil {
		_, _ = w.Write([]byte(err.Error()))
		return
	}
	b := bufio.NewReader(np)
	item, err := cfg.Storage.CreateItem(r.PathValue("share"), np.FileName(), b)
	if err != nil {
		w.WriteHeader(http.StatusNotAcceptable)
		_, _ = w.Write([]byte(err.Error()))
		return
	}

	c, err := json.Marshal(item)
	if err != nil {
		_, _ = w.Write([]byte(err.Error()))
		return
	}
	_, _ = w.Write(c)
}

// getShares returns the list of shares as json
func getShares(w http.ResponseWriter, r *http.Request) {
	shares, err := cfg.Storage.ListShares()

	if err != nil {
		slog.Error("getShares", slog.String("error", err.Error()))
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(err.Error()))
		return
	}
	response, err := json.Marshal(shares)
	if err != nil {
		slog.Error("getShares", slog.String("error", err.Error()))
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(err.Error()))
		return
	}
	_, _ = w.Write(response)
}

// getShare returns the share identified by the request parameter
func getShare(w http.ResponseWriter, r *http.Request) {
	share, err := cfg.Storage.GetShare(r.PathValue("share"))
	if err != nil {
		slog.Error("getShares", slog.String("error", err.Error()))
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(err.Error()))
		return
	}

	if auth.UserForRequest(r) == "" && !share.IsValid() {
		w.WriteHeader(http.StatusGone)
		_, _ = w.Write([]byte("Share expired"))
		return
	}

	content, err := cfg.Storage.ListShare(r.PathValue("share"))
	if err != nil {
		slog.Error("getShares", slog.String("error", err.Error()))
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(err.Error()))
		return
	}
	b, err := json.Marshal(content)
	if err != nil {
		slog.Error("getShares", slog.String("error", err.Error()))
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(err.Error()))
		return
	}
	_, _ = w.Write(b)
}

// deleteShare deletes the share identified by the request parameter
func deleteShare(w http.ResponseWriter, r *http.Request) {
	err := cfg.Storage.DeleteShare(r.PathValue("share"))
	if err != nil {
		slog.Error("deleteShare", slog.String("error", err.Error()))
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(err.Error()))
		return
	}
}

// getItem returns the item identified by the request parameter
func getItem(w http.ResponseWriter, r *http.Request) {
	shareName := r.PathValue("share")
	itemName := r.PathValue("item")

	item, err := cfg.Storage.GetItem(shareName, itemName)
	if err != nil {
		slog.Error("getItem", slog.String("error", err.Error()))
		_, _ = w.Write([]byte(err.Error()))
		return
	}

	reader, err := cfg.Storage.GetItemData(shareName, itemName)
	if err != nil {
		slog.Error("getShares", slog.String("error", err.Error()))
		_, _ = w.Write([]byte(err.Error()))
		return
	}

	defer reader.Close()

	w.Header().Add("Content-Type", "application/octet-stream")
	w.Header().Add("Content-Length", fmt.Sprintf("%d", item.ItemInfo.Size))
	w.Header().Add("Content-Disposition", "attachment")

	_, err = io.Copy(w, reader)
	if err != nil {
		slog.Error("getItem", slog.String("error", err.Error()))
		_, _ = w.Write([]byte(err.Error()))
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

	b, err := json.Marshal(u)
	if err != nil {
		slog.Error("postLogin", slog.String("error", err.Error()))
		w.WriteHeader(http.StatusInternalServerError)
	}

	_, _ = w.Write(b)
}

// getVersion returns hupload version
func getVersion(w http.ResponseWriter, r *http.Request) {
	v := struct {
		Version string `json:"version"`
	}{
		Version: version,
	}

	b, err := json.Marshal(v)
	if err != nil {
		slog.Error("getVersion", slog.String("error", err.Error()))
		w.WriteHeader(http.StatusInternalServerError)
	}

	_, _ = w.Write(b)
}
