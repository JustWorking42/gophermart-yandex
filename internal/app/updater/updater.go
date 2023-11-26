package updater

import (
	"context"
	"encoding/json"
	"sync"
	"time"

	"github.com/JustWorking42/gophermart-yandex/internal/app/model"
	"github.com/JustWorking42/gophermart-yandex/internal/app/repository"
	"github.com/go-resty/resty/v2"
)

type Updater struct {
	repository  *repository.AppRepository
	serviceAddr string
}

func NewUpdater(repository *repository.AppRepository, serviceAddr string) *Updater {
	return &Updater{
		repository:  repository,
		serviceAddr: serviceAddr,
	}
}

func (u *Updater) SubcribeOnTask(ctx context.Context) (*sync.WaitGroup, chan error) {
	ticker := time.NewTicker(time.Second * 10)
	var wg sync.WaitGroup
	errChan := make(chan error)
	wg.Add(1)
	go func() {
		for {
			select {
			case <-ticker.C:
				orders, err := u.repository.GetNonProcessedOrdersID(ctx)
				if err != nil {
					errChan <- err
					continue
				}

				if len(orders) > 0 {
					u.updateOrderStatus(orders)

					if err != nil {
						errChan <- err
						continue
					}
				}

			case <-ctx.Done():
				wg.Done()
				close(errChan)
				return
			}
		}
	}()
	return &wg, errChan
}

func (u *Updater) updateOrderStatus(orders []string) {
	client := resty.New()
	var wg sync.WaitGroup

	for _, order := range orders {
		wg.Add(1)
		go func(order string) {
			defer wg.Done()

			resp, err := client.R().Get(u.serviceAddr + "/api/orders/" + order)
			if err != nil {
				return
			}

			var orderModel model.OrderModel
			err = json.Unmarshal(resp.Body(), &orderModel)
			if err != nil {
				return
			}

			err = u.repository.UpdateOrderStatus(context.Background(), order, float64(orderModel.Accrual), orderModel.Status)
			if err != nil {
				return
			}
		}(order)
	}

	wg.Wait()
}
