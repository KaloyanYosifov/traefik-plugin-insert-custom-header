# Mutate Headers

Insert Custom Header is a middleware plugin for [Traefik](https://traefik.io) that can create new headers from existing headers or URL

## Configuration

### Fields

#### Mutations

- `header` (string, required): The name of the header to be mutated.
- `newName` (string, optional): The new name of the header. If not provided, the header will be mutated in place.
- `regex` (string, optional): The regular expression to match the header value. If not set the header value will be preserved.
- `replacement` (string, optional): The replacement string for the header value. Must be set if `regex` is set.

#### FromUrlMutations
- `header` (string, required): The name of the header to be mutated.
- `newName` (string, optional): The new name of the header. If not provided, the header will be mutated in place.
- `regex` (string, optional): The regular expression to match the header value. If not set the header value will be preserved.
- `replacement` (string, optional): The replacement string for the header value. Must be set if `regex` is set.

### Static

```yaml
experimental:
  plugins:
    traefik-plugin-mutate-headers:
      modulename: "https://github.com/KaloyanYosifov/traefik-plugin-insert-custom-header"
      version: "v1.0.0"
```
