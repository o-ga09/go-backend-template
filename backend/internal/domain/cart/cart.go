package cart

import (
	"github.com/o-ga09/go-backend-template/pkg/authz"
	"github.com/o-ga09/go-backend-template/pkg/errors"
)

func (c *Cart) IsOwnedBy(requesterID string) bool {
	return authz.IsOwner(requesterID, c.UserID)
}

// sentinelを返すのは呼び出し元がerrors.Isで判別できるようにするため。
func (c *Cart) CanCheckout() error {
	if len(c.Items) == 0 {
		return errors.ErrCartEmpty
	}
	return nil
}
