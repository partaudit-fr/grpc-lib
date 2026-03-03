package apperror

// defaultMessages contains user-friendly messages for the generic error codes.
var defaultMessages = map[ErrorCode]string{
	INTERNAL_ERROR:    "Une erreur interne est survenue. Veuillez réessayer.",
	UNAUTHORIZED:      "Vous devez être connecté pour effectuer cette action.",
	FORBIDDEN:         "Vous n'avez pas les droits pour effectuer cette action.",
	NOT_FOUND:         "La ressource demandée est introuvable.",
	VALIDATION_FAILED: "Les données fournies sont invalides.",
	CONFLICT:          "Un conflit est survenu. Veuillez réessayer.",
}
