package profile

import (
	"time"

	"github.com/Southclaws/dt"
	"github.com/Southclaws/fault"
	"github.com/Southclaws/opt"
	"github.com/rs/xid"

	"github.com/Southclaws/storyden/app/resources/account"
	"github.com/Southclaws/storyden/app/resources/account/role/held"
	"github.com/Southclaws/storyden/app/resources/asset"
	"github.com/Southclaws/storyden/app/resources/datagraph"
	"github.com/Southclaws/storyden/app/resources/tag/tag_ref"
	"github.com/Southclaws/storyden/internal/ent"
)

type Ref struct {
	ID       account.AccountID
	Created  time.Time
	Updated  time.Time
	Deleted  opt.Optional[time.Time]
	Handle   string
	Name     string
	Bio      datagraph.Content
	Admin    bool
	Metadata map[string]any
}

func MapRef(a *ent.Account) (*Ref, error) {
	return &Ref{
		ID:      account.AccountID(a.ID),
		Created: a.CreatedAt,
		Updated: a.UpdatedAt,
		Deleted: opt.NewPtr(a.DeletedAt),
		Handle:  a.Handle,
		Name:    a.Name,
	}, nil
}

type Public struct {
	Ref

	Followers     int
	Following     int
	LikeScore     int
	Roles         held.Roles
	Interests     []*tag_ref.Tag
	ExternalLinks []account.ExternalLink
	InvitedBy     opt.Optional[Ref]
}

func (p *Public) GetID() xid.ID                 { return xid.ID(p.ID) }
func (p *Public) GetKind() datagraph.Kind       { return datagraph.KindProfile }
func (p *Public) GetName() string               { return p.Name }
func (p *Public) GetSlug() string               { return p.Handle }
func (p *Public) GetDesc() string               { return p.Bio.Short() }
func (p *Public) GetContent() datagraph.Content { return p.Bio }
func (p *Public) GetProps() map[string]any      { return p.Metadata }
func (p *Public) GetAssets() []*asset.Asset     { return []*asset.Asset{} }
func (p *Public) GetCreated() time.Time         { return p.Created }
func (p *Public) GetUpdated() time.Time         { return p.Updated }

func Map(a *ent.Account) (*Public, error) {
	rolesEdge, err := a.Edges.AccountRolesOrErr()
	if err != nil {
		return nil, err
	}

	tagsEdge, err := a.Edges.TagsOrErr()
	if err != nil {
		return nil, err
	}

	roles, err := held.MapList(rolesEdge, a.Admin)
	if err != nil {
		return nil, fault.Wrap(err)
	}

	interests := dt.Map(tagsEdge, tag_ref.Map(nil))

	bio, err := datagraph.NewRichText(a.Bio)
	if err != nil {
		return nil, err
	}

	invitedByEdge := opt.NewPtr(a.Edges.InvitedBy)

	invitedBy, err := opt.MapErr(invitedByEdge, func(i ent.Invitation) (Ref, error) {
		c, err := i.Edges.CreatorOrErr()
		if err != nil {
			return Ref{}, err
		}

		p, err := MapRef(c)
		if err != nil {
			return Ref{}, err
		}

		return *p, nil
	})
	if err != nil {
		return nil, err
	}

	links, err := dt.MapErr(a.Links, account.MapExternalLink)
	if err != nil {
		return nil, fault.Wrap(err)
	}

	return &Public{
		Ref: Ref{
			ID:       account.AccountID(a.ID),
			Created:  a.CreatedAt,
			Updated:  a.UpdatedAt,
			Deleted:  opt.NewPtr(a.DeletedAt),
			Handle:   a.Handle,
			Name:     a.Name,
			Bio:      bio,
			Metadata: a.Metadata,
		},
		Followers: 0, // TODO: Hydrate here
		Following: 0, // TODO: Hydrate here
		LikeScore: 0, // TODO: Hydrate here

		Roles:         roles,
		Interests:     interests,
		InvitedBy:     invitedBy,
		ExternalLinks: links,
	}, nil
}
