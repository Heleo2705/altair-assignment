package todolist

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"go.altair.com/todolist/pkg/structs"
)

const (
	MediaTypeJSON = "application/json"
)

type ItemsHandlers struct {
	ItemsService ItemsService
}

func (h *ItemsHandlers) ConfigureRoutes(r chi.Router) {
	r.Route("/todolist", func(r chi.Router) {
		r.Post("/", h.createItem)
		r.Get("/", h.listItems)

		r.Route("/{id}", func(r chi.Router) {
			r.Get("/", h.getItem)
			r.Put("/", h.updateItem)
			r.Put("/order", h.updateOrder)
			r.Delete("/", h.deleteItem)
		})
	})
}

func requestAs(r *http.Request, v interface{}) error {
	if r.ContentLength == 0 {
		return nil
	} else { // assume JSON by default
		decoder := json.NewDecoder(r.Body)
		if err := decoder.Decode(v); err != nil {
			return err
		}
	}
	return nil
}

type CreateItemRequest struct {
	Item     structs.TodoItem
	ListSize int
}

func (h *ItemsHandlers) createItem(w http.ResponseWriter, r *http.Request) {
	var Request CreateItemRequest
	err := requestAs(r, &Request)
	if err != nil {
		http.Error(w, "Failed", http.StatusBadRequest)
		return
	}
	Request.Item.ItemOrder = Request.ListSize + 1
	err = h.ItemsService.AddItem(r.Context(), &Request.Item)
	if err != nil {
		http.Error(w, "Failed", http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusAccepted)
}

func (h *ItemsHandlers) listItems(w http.ResponseWriter, r *http.Request) {
	items, err := h.ItemsService.ListItems(r.Context())
	if err != nil {
		http.Error(w, "Failed", http.StatusBadRequest)
		return
	}

	w.Header().Add("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(items)
}

func (h *ItemsHandlers) deleteItem(w http.ResponseWriter, r *http.Request) {
	deploymentId := chi.URLParam(r, "id")
	err := h.ItemsService.DeleteItem(r.Context(), deploymentId)
	if err != nil {
		http.Error(w, "Failed", http.StatusBadRequest)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *ItemsHandlers) updateItem(w http.ResponseWriter, r *http.Request) {
	deploymentId := chi.URLParam(r, "id")

	var item structs.TodoItem
	err := requestAs(r, &item)
	if err != nil {
		http.Error(w, "Failed", http.StatusBadRequest)
		return
	}

	item.Id = deploymentId

	err = h.ItemsService.UpdateItem(r.Context(), &item)
	if err != nil {
		http.Error(w, "Failed", http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusAccepted)
}

func (h *ItemsHandlers) getItem(w http.ResponseWriter, r *http.Request) {
	deploymentId := chi.URLParam(r, "id")

	deployment, err := h.ItemsService.GetItem(r.Context(), deploymentId)
	if err != nil {
		http.Error(w, "Failed", http.StatusBadRequest)
		return
	}

	w.Header().Add("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(deployment)
}

type UpdateOrderRequest struct {
	Start int
	End   int
}

func (h *ItemsHandlers) updateOrder(w http.ResponseWriter, r *http.Request) {
	itemId := chi.URLParam(r, "id")

	item, err := h.ItemsService.GetItem(r.Context(), itemId)
	if err != nil {
		http.Error(w, "Failed", 401)
		return
	}
	var request UpdateOrderRequest
	err = requestAs(r, &request)
	if err != nil {
		http.Error(w, "Failed", 402)
		return
	}
	listItems, err := h.ItemsService.ListItems(r.Context())
	if err != nil {
		http.Error(w, "Failed", 403)
		return
	}
	if request.Start < request.End {

		for i := range listItems.Items {
			curr := listItems.Items[i]
			if curr.ItemOrder < request.End && curr.ItemOrder > request.Start {

				curr.ItemOrder = curr.ItemOrder - 1
				err = h.ItemsService.UpdateItem(r.Context(), &curr)
				if err != nil {
					http.Error(w, "Failed", 404)
					return
				}
			}
		}
		item.ItemOrder = request.End + 1
		err = h.ItemsService.UpdateItem(r.Context(), item)
		if err != nil {
			http.Error(w, "Failed", 405)
			return
		}
	} else if request.Start > request.End {
		for i := range listItems.Items {
			curr := listItems.Items[i]
			if curr.ItemOrder > request.End && curr.ItemOrder < request.Start {
				curr.ItemOrder = curr.ItemOrder + 1
				err = h.ItemsService.UpdateItem(r.Context(), &curr)
				if err != nil {
					http.Error(w, "Failed", 406)
					return
				}
			}
		}
		item.ItemOrder = request.End + 1
		err = h.ItemsService.UpdateItem(r.Context(), item)
		if err != nil {
			http.Error(w, "Failed", 407)
			return
		}
	}
	w.Header().Add("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(item)
}
