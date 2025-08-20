// Package docs expõe os arquivos de documentação OpenAPI/Swagger embutidos.
// Mantemos um swagger.yaml estático versionado e servimos via HTTP.
// Para editar a especificação modifique swagger.yaml diretamente.
package docs

import "embed"

//go:embed swagger.yaml swagger.json
var SwaggerFS embed.FS
