# Interfaces Go — Phase 2 Profils

## Nouveaux ports

### `ports.MemberRepository`

```go
type MemberRepository interface {
    Create(ctx context.Context, m *member.Member) error
    FindByAccountID(ctx context.Context, accountID string) (*member.Member, error)
    Update(ctx context.Context, m *member.Member) error
}
```

### `ports.AssociationRepository`

```go
type AssociationRepository interface {
    Create(ctx context.Context, a *association.Association) error
    FindByAccountID(ctx context.Context, accountID string) (*association.Association, error)
    Update(ctx context.Context, a *association.Association) error
}
```

### `ports.MembershipRepository`

```go
type MembershipRepository interface {
    Create(ctx context.Context, m *membership.Membership) error
    FindByAccountIDs(ctx context.Context, memberID, associationID string) (*membership.Membership, error)
}
```

### `ports.FileStore`

```go
type FileStore interface {
    // StoreAvatar stores the original file and returns its path relative to data/uploads/.
    StoreAvatar(ctx context.Context, accountID string, content io.Reader, ext string) (path string, err error)
    // AvatarURL returns the public URL for a given avatar path.
    AvatarURL(path string) string
}
```

## Ports modifiés

### `ports.AccountRepository` (ajout)

```go
// Existing interface — add:
FindByID(ctx context.Context, id string) (*account.Account, error)
UpdateKindAndStatus(ctx context.Context, id string, kind account.Kind, status account.Status) error
```

## Nouveaux use cases

### `application/auth.RegisterMemberService`

```go
func (s *RegisterMemberService) RegisterMember(ctx context.Context, ip string, in RegisterMemberInput) error
// RegisterMemberInput: Email, Password, FirstName, LastName, BirthDate
```

### `application/auth.RegisterAssociationService`

```go
func (s *RegisterAssociationService) RegisterAssociation(ctx context.Context, ip string, in RegisterAssociationInput) error
// RegisterAssociationInput: Email, Password, LegalName, PostalCode, SIREN, RNA, FirstName, LastName, BirthDate
```

### `application/profile.GetProfileService`

```go
func (s *GetProfileService) GetMyProfile(ctx context.Context, sessionID string) (ProfileOutput, error)
func (s *GetProfileService) GetPublicProfile(ctx context.Context, accountID string, viewerSessionID string) (ProfileOutput, error)
// ProfileOutput: discriminated union via Kind field
```

### `application/profile.UpdateProfileService`

```go
func (s *UpdateProfileService) UpdateProfile(ctx context.Context, sessionID string, in UpdateProfileInput) error
// UpdateProfileInput: Kind-aware patch struct
```

### `application/profile.UploadAvatarService`

```go
func (s *UploadAvatarService) UploadAvatar(ctx context.Context, sessionID string, r io.Reader, size int64, mimeType string) (avatarURL string, err error)
```
