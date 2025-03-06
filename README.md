# OpenTelemetry HTML Viewer

Una herramienta en Go que genera visualizaciones HTML de trazas OpenTelemetry y las compara en Pull Requests de GitHub.

## Características

- Lee archivos JSON con trazas OpenTelemetry
- Genera visualizaciones HTML interactivas
- Compara trazas entre diferentes versiones
- Comenta automáticamente en PRs de GitHub

## Instalación

```bash
go install github.com/lcalisi/otel-html-viewer@latest
```

## Uso

### Básico

```bash
otel-html-viewer -i traces.json -o visualization.html
```

### Con integración de GitHub

```bash
otel-html-viewer \
  -i traces.json \
  -o visualization.html \
  -t $GITHUB_TOKEN \
  -p 123 \
  --owner tu-usuario \
  --repo tu-repositorio
```

## Parámetros

- `-i, --input`: Archivo JSON con las trazas OpenTelemetry (requerido)
- `-o, --output`: Archivo HTML de salida (requerido)
- `-t, --github-token`: Token de GitHub para comentarios en PRs (opcional)
- `-p, --pr-number`: Número del PR para comentar (opcional)
- `--owner`: Propietario del repositorio (opcional)
- `--repo`: Nombre del repositorio (opcional)

## Formato del JSON de entrada

El archivo JSON de entrada debe contener un array de trazas en el siguiente formato:

```json
[
  {
    "trace_id": "string",
    "spans": [
      {
        "span_id": "string",
        "parent_span_id": "string",
        "name": "string",
        "start_time": "timestamp",
        "end_time": "timestamp",
        "attributes": {
          "key": "value"
        },
        "events": [
          {
            "time": "timestamp",
            "name": "string",
            "attributes": {
              "key": "value"
            }
          }
        ]
      }
    ],
    "start_time": "timestamp",
    "end_time": "timestamp",
    "attributes": {
      "key": "value"
    }
  }
]
```

## Contribuir

Las contribuciones son bienvenidas. Por favor, abre un issue para discutir los cambios que te gustaría hacer. 