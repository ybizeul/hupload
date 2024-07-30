package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"

	"github.com/ybizeul/hupload/pkg/apiws"
)

func postShare(w http.ResponseWriter, r *http.Request) {
	user := apiws.UserForRequest(r)
	if user == "" {
		slog.Error("putShare", slog.String("error", "no user in context"))
		_, _ = w.Write([]byte("no user in context"))
		return
	}
	code := generateCode()
	err := api.StorageService.CreateShare(code, user)
	if err != nil {
		slog.Error("putShare", slog.String("error", err.Error()))
		_, _ = w.Write([]byte(err.Error()))
		return
	}
	_, _ = w.Write([]byte(code))

}

func putShare(w http.ResponseWriter, r *http.Request) {
	user := apiws.UserForRequest(r)
	if user == "" {
		slog.Error("putShare", slog.String("error", "no user in context"))
		_, _ = w.Write([]byte("no user in context"))
		return
	}
	err := api.StorageService.CreateShare(r.PathValue("share"), user)
	if err != nil {
		slog.Error("putShare", slog.String("error", err.Error()))
		_, _ = w.Write([]byte(err.Error()))
		return
	}
}

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
	err = api.StorageService.CreateItem(r.PathValue("share"), np.FileName(), b)
	if err != nil {
		_, _ = w.Write([]byte(err.Error()))
		return
	}
	_, _ = w.Write([]byte("OK"))
}

func getShares(w http.ResponseWriter, r *http.Request) {
	shares, err := api.StorageService.ListShares()
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

func getShare(w http.ResponseWriter, r *http.Request) {
	content, err := api.StorageService.ListShare(r.PathValue("share"))
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

func getItem(w http.ResponseWriter, r *http.Request) {
	item, err := api.StorageService.GetItem(r.PathValue("share"), r.PathValue("item"))
	if err != nil {
		slog.Error("getShares", slog.String("error", err.Error()))
		_, _ = w.Write([]byte(err.Error()))
		return
	}
	reader, err := api.StorageService.GetItemData(r.PathValue("share"), r.PathValue("item"))
	if err != nil {
		slog.Error("getShares", slog.String("error", err.Error()))
		_, _ = w.Write([]byte(err.Error()))
		return
	}
	w.Header().Add("Content-Length", fmt.Sprintf("%d", item.ItemInfo.Size))
	_, _ = io.Copy(w, reader)
}

func postLogin(w http.ResponseWriter, r *http.Request) {
	_, _ = w.Write([]byte("OK"))
}
