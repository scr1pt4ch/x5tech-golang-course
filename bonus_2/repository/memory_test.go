package repository_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"example.com/delivery/repository"
	"example.com/delivery/service"
)

func TestDefaultFindsAllTariffs(t *testing.T) {
	t.Parallel()

	repo := repository.Default()

	for _, name := range []string{"courier", "pickup", "express"} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			pricer, err := repo.Find(name)

			require.NoError(t, err)
			assert.NotNil(t, pricer)
		})
	}
}

func TestFindUnknownTariff(t *testing.T) {
	t.Parallel()

	pricer, err := repository.Default().Find("drone")

	assert.ErrorIs(t, err, service.ErrUnknownTariff)
	assert.Nil(t, pricer)
}

func TestNewMemoryCopiesMap(t *testing.T) {
	t.Parallel()

	pricers := map[string]service.DeliveryPricer{"courier": service.NewCourier()}
	repo := repository.NewMemory(pricers)

	delete(pricers, "courier")
	pricers["drone"] = service.NewExpress()

	_, err := repo.Find("courier")
	assert.NoError(t, err)

	_, err = repo.Find("drone")
	assert.ErrorIs(t, err, service.ErrUnknownTariff)
}
