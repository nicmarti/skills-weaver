package campaign

import (
	"errors"
	"fmt"

	"dungeons/internal/llm"
)

// FriendlyError translates a campaign-plan generation failure into a
// French, user-facing message suitable for the web UI and the CLI. Technical
// details stay attached via errors.Unwrap for logs.
func FriendlyError(err error) string {
	var llmErr *llm.Error
	switch {
	case errors.As(err, &llmErr):
		switch llmErr.Kind {
		case llm.ErrKindTimeout:
			return "La génération du plan de campagne a dépassé le temps imparti (5 minutes). " +
				"Aucune aventure n'a été créée — réessaie, ou choisis un modèle plus rapide via OPENROUTER_MODEL_CAMPAIGN."
		case llm.ErrKindAuth:
			return "Authentification OpenRouter refusée (clé invalide). Aucune aventure n'a été créée — vérifie OPENROUTER_API_KEY."
		case llm.ErrKindPayment:
			return "Crédits OpenRouter insuffisants pour générer le plan de campagne. Aucune aventure n'a été créée."
		case llm.ErrKindRateLimit:
			return "Limite de requêtes atteinte côté OpenRouter ou fournisseur. Aucune aventure n'a été créée — réessaie dans quelques instants."
		case llm.ErrKindOverloaded, llm.ErrKindUnavailable:
			return "Le fournisseur du modèle est momentanément surchargé ou indisponible. Aucune aventure n'a été créée — réessaie."
		case llm.ErrKindCanceled:
			return "La génération du plan de campagne a été annulée. Aucune aventure n'a été créée."
		case llm.ErrKindMidStream:
			return "La connexion au modèle a été interrompue en cours de génération. Aucune aventure n'a été créée — réessaie."
		default:
			return fmt.Sprintf("Le modèle n'a pas pu générer le plan de campagne : %s. Aucune aventure n'a été créée.",
				llmErr.Error())
		}
	default:
		return fmt.Sprintf("Échec de la génération du plan de campagne : %v. Aucune aventure n'a été créée.", err)
	}
}
