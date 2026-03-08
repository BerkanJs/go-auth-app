package service

// AuthorizationStrategy, bir kaynağa erişim iznini belirleyen Strategy arayüzü.
// Farklı yetkilendirme kuralları bu arayüzü implemente ederek birbirinin yerine kullanılabilir.
type AuthorizationStrategy interface {
	IsAuthorized(userRole string, userID int, resourceOwnerID int) bool
}

// AdminOnlyStrategy yalnızca admin rolüne erişim izni verir.
type AdminOnlyStrategy struct{}

func (s *AdminOnlyStrategy) IsAuthorized(userRole string, userID int, resourceOwnerID int) bool {
	return userRole == "admin"
}

// OwnerOrAdminStrategy, kaynağın sahibine veya admin rolüne erişim izni verir.
// Blog güncelleme ve silme işlemlerinde kullanılır.
type OwnerOrAdminStrategy struct{}

func (s *OwnerOrAdminStrategy) IsAuthorized(userRole string, userID int, resourceOwnerID int) bool {
	return userRole == "admin" || userID == resourceOwnerID
}
