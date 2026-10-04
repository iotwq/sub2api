//go:build unit

package service

import (
	"context"
	"maps"
	"testing"

	"github.com/stretchr/testify/require"
)

type adminBasispointsGroupRepo struct {
	accountRepoStubForBulkUpdate
	extraUpdates []map[string]any
}

func (r *adminBasispointsGroupRepo) UpdateExtra(_ context.Context, _ int64, updates map[string]any) error {
	r.extraUpdates = append(r.extraUpdates, maps.Clone(updates))
	return nil
}

func TestAdminBasispoints403GroupValidationBeforeWrite(t *testing.T) {
	for _, method := range []string{"create", "update", "extra", "bulk"} {
		for _, valid := range []bool{false, true} {
			if method == "create" && valid {
				continue // Successful OAuth creation starts unrelated asynchronous privacy setup.
			}
			t.Run(method+map[bool]string{false: "/missing target", true: "/explicit leave all"}[valid], func(t *testing.T) {
				account := basispointsAccountForTest()
				repo := &adminBasispointsGroupRepo{accountRepoStubForBulkUpdate: accountRepoStubForBulkUpdate{
					getByIDAccounts: map[int64]*Account{account.ID: account}, getByIDsAccounts: []*Account{account},
				}}
				svc := &adminServiceImpl{accountRepo: repo}
				extra := map[string]any{BasispointsAutoMoveOn403Key: true}
				if valid {
					extra[Basispoints403TargetGroupIDKey] = float64(0)
				}
				ctx := context.Background()
				var err error
				switch method {
				case "create":
					// Invalid input must be rejected before persistence or OAuth side effects.
					_, err = svc.CreateAccount(ctx, &CreateAccountInput{Name: "bps-policy", Platform: PlatformOpenAI, Type: AccountTypeOAuth,
						Credentials: account.Credentials, Extra: extra, SkipDefaultGroupBind: true})
				case "update":
					_, err = svc.UpdateAccount(ctx, account.ID, &UpdateAccountInput{Extra: extra})
				case "extra":
					err = svc.UpdateAccountExtra(ctx, account.ID, extra)
				case "bulk":
					_, err = svc.BulkUpdateAccounts(ctx, &BulkUpdateAccountsInput{AccountIDs: []int64{account.ID}, Extra: extra})
				}
				if valid {
					require.NoError(t, err)
					require.Equal(t, 1, len(repo.updatedAccounts)+len(repo.extraUpdates)+repo.bulkUpdateCalls)
				} else {
					requireApplicationErrorReason(t, err, "OPENAI_BASISPOINTS_INVALID")
					require.Nil(t, repo.createAccount)
					require.Empty(t, repo.updatedAccounts)
					require.Empty(t, repo.extraUpdates)
					require.Zero(t, repo.bulkUpdateCalls)
				}
			})
		}
	}
}

func TestAdminBasispoints403GroupRejectsNullTargetBeforeWrite(t *testing.T) {
	account := basispointsAccountForTest()
	account.Extra[BasispointsAutoMoveOn403Key] = true
	account.Extra[Basispoints403TargetGroupIDKey] = float64(7)
	repo := &adminBasispointsGroupRepo{accountRepoStubForBulkUpdate: accountRepoStubForBulkUpdate{
		getByIDAccounts: map[int64]*Account{account.ID: account},
	}}
	svc := &adminServiceImpl{accountRepo: repo}
	err := svc.UpdateAccountExtra(context.Background(), account.ID, map[string]any{Basispoints403TargetGroupIDKey: nil})
	requireApplicationErrorReason(t, err, "OPENAI_BASISPOINTS_INVALID")
	require.Empty(t, repo.extraUpdates)
	require.Equal(t, float64(7), account.Extra[Basispoints403TargetGroupIDKey])
}

func TestAdminBasispoints403GroupPartialUpdateKeepsExplicitTarget(t *testing.T) {
	for _, method := range []string{"extra", "bulk"} {
		account := basispointsAccountForTest()
		account.Extra[Basispoints403TargetGroupIDKey] = float64(0)
		repo := &adminBasispointsGroupRepo{accountRepoStubForBulkUpdate: accountRepoStubForBulkUpdate{
			getByIDAccounts: map[int64]*Account{account.ID: account}, getByIDsAccounts: []*Account{account},
		}}
		svc := &adminServiceImpl{accountRepo: repo}
		extra := map[string]any{BasispointsAutoMoveOn403Key: true}
		var err error
		if method == "extra" {
			err = svc.UpdateAccountExtra(context.Background(), account.ID, extra)
		} else {
			_, err = svc.BulkUpdateAccounts(context.Background(), &BulkUpdateAccountsInput{AccountIDs: []int64{account.ID}, Extra: extra})
		}
		require.NoError(t, err)
		require.NotContains(t, account.Extra, BasispointsAutoMoveOn403Key, "validation cannot mutate shared account data")
	}
}
