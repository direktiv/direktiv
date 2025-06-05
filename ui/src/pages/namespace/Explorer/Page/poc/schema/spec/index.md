# Page Document Spec

## Principles

- Page documents have no knowledge of the Direktiv Gateway. API requests are defined as generic fetch requests. However the UI can help the user pick a proper endpoint by suggesting routes that are available in the current environment. This principle makes a page document more portable and independent of the gateway's specification. This means if a user updates a gateway's endpoint, they also have to update all potential pages that use these endpoints.
- To keep the specification simple, we should not introduce any interface to configure the theme of the page yet. In a future version, we might consider having a global theme setting that could then be injected to override CSS variables. This might require an update to Tailwind 4.
- The specification by design allows for the composition of sections that "don't make sense" and can make the page look or behave strangely, similar to HTML when you combine tags in an invalid way, like a button inside a button or a div inside a p tag. However, we will design the user interface in a way that prevents the user from doing things they should not do, such as placing a modal inside another modal.

## Direktiv Page

| Attribute      | Type         | Description                      |
| -------------- | ------------ | -------------------------------- |
| `direktiv_api` | `"pages/v1"` |                                  |
| `blocks`       | `Block[]`    | Entry point for all page content |

# Blocks

Blocks are the main elements that the user can use to compose a Direktiv page.

## Headline `Block`

| Attribute | Type                   | Description           |
| --------- | ---------------------- | --------------------- |
| `type`    | `"headline"`           |                       |
| `label`   | `TemplateString`       | Main headline         |
| `level`   | `"h1"`, `"h2"`, `"h3"` | Level of the headline |

### Text `Block`

| Attribute | Type             | Description  |
| --------- | ---------------- | ------------ |
| `type`    | `"text"`         |              |
| `content` | `TemplateString` | Text content |

### Card `Block`

| Attribute | Type      | Description  |
| --------- | --------- | ------------ |
| `type`    | `"card"`  |              |
| `blocks`  | `Block[]` | Card content |

### Columns `Block`

| Attribute | Type           | Description                                        |
| --------- | -------------- | -------------------------------------------------- |
| `type`    | `"columns"`    |                                                    |
| `blocks`  | `Column[]`     | Array of column blocks                             |

### Column `Block`

| Attribute | Type       | Description                    |
| --------- | ---------- | ------------------------------ |
| `type`    | `"column"` |                                |
| `blocks`  | `Block[]`  | Content blocks within the column |

### Dialog `Block`

| Attribute | Type             | Description                     |
| --------- | ---------------- | ------------------------------- |
| `type`    | `"dialog"`       |                                 |
| `trigger` | `Block<trigger>` | Opens the dialog when clicked   |
| `blocks`  | `Block[]`        | Content shown inside the dialog |

### Table `Block`

| Attribute | Type                   | Description                                               |
| --------- | ---------------------- | --------------------------------------------------------- |
| `type`    | `"table"`              |                                                           |
| `data`    | `Block<Loop>`          | the loop block to interate on                             |
| `actions` | `Block<Button>[]`      | List of actions that will be available in the last column |
| `columns` | `Block<TableColumn>[]` | List of table columns                                     |

### TableColumn `Block`

| Attribute | Type             | Description            |
| --------- | ---------------- | ---------------------- |
| `type`    | `"table-column"` |                        |
| `label`   | `TemplateString` | Headline of the column |
| `content` | `TemplateString` | Cell content           |

### Button `Block<trigger>`

| Attribute | Type       | Description                                                                                                        |
| --------- | ---------- | ------------------------------------------------------------------------------------------------------------------ |
| `type`    | `"button"` |                                                                                                                    |
| `label`   | `string`   | Button label text                                                                                                  |
| `submit?` | `Mutation` | mutation that will be executed on click. This is optional as the button can also be used as a trigger for a dialog |

### Form `Block`

| Attribute  | Type             | Description                      |
| ---------- | ---------------- | -------------------------------- |
| `type`     | `"form"`         |                                  |
| `trigger`  | `Block<trigger>` | Trigger block to submit the form |
| `mutation` | `Mutation`       | Mutation executed on submission  |
| `blocks`   | `Block[]`        | Form content                     |

### QueryProvider `Block`

A QueryProvider is a block that is responsible for fetching data from one or multiple APIs and providing the result to its child blocks. It will display a loading indicator until all data is fetched successfully. Every child block can access the data in fields of type `TemplateString` by using the `id` of the corresponding `Query` inside the `QueryProvider` as a reference.

| Attribute | Type               | Description                      |
| --------- | ------------------ | -------------------------------- |
| `type`    | `"query-provider"` |                                  |
| `queries` | `Query[]`          | Queries to execute               |
| `blocks`  | `Block[]`          | Children with access to the data |

### Loop `Block`

A Block that allows to iterate over an array variable. It renders its blocks for each item.

| Attribute | Type              | Description                    |
| --------- | ----------------- | ------------------------------ |
| `type`    | `"loop"`          |                                |
| `id`      | `Id`              | Unique identifier              |
| `data`    | `Variable<Array>` | Variable to loop over          |
| `blocks`  | `Block[]`         | Child blocks rendered per item |

**Example**

When adding a `loop` block with the id `pokemon` and a variable `query.pokemonList.data` (points to a `query` with the id `pokemon-list` and references the `data` array returned from the API request), every block will have access to the variable `loop.pokemon` to reference its corresponding array item.

# Procedures

Procedures are types that can be used in various blocks that will handle API requests

### `Mutation`

A mutation is an API request that modifies data on the server

| Attribute         | Type             | Description               |
| ----------------- | ---------------- | ------------------------- |
| `id`              | `Id`             | Unique identifier         |
| `method`          | `MutationMethod` | HTTP method               |
| `endpoint`        | `TemplateString` | API endpoint              |
| `queryParams?`    | `KeyValue[]`     | Optional query parameters |
| `requestHeaders?` | `KeyValue[]`     | Optional request headers  |
| `requestBody?`    | `KeyValue[]`     | Optional request body     |

### `MutationMethod`

`"POST"`, `"PUT"`, `"PATCH"`, `"DELETE"`

## `Query`

A query is a API request that reads data from the server

| Attribute      | Type             | Description                            |
| -------------- | ---------------- | -------------------------------------- |
| `id`           | `Id`             | Unique ID used to reference query data |
| `endpoint`     | `TemplateString` | Path to the endpoint                   |
| `queryParams?` | `KeyValue[]`     | Optional query parameters              |

# Primitives

## `Variable<String | Boolean | Array>`

A Variable can be sourced from special parent blocks like e.g. a `form` or a `query`.

**Examples**

- `query.pokemon.data.name`
  This looks for a parent `query` block, with the name `pokemon` and points to the `data.name` attribute of the JSON response of that request

_\*the exact syntax is still TBD_

## `TemplateString`

A template string is a string that can have `Variable` placeholders that will be filled with dynamic data. Variables will always be stringified.

**Example**s

- `Edit {{query.pokemon.data.name}}`

_\*the exact syntax is still TBD_

## `Id`

An Id is a string that is unique within a page and identifies a resource. IDs are used when one resource needs to reference another resource, like when one block references dynamic data from a query.

## `KeyValue`

| Attribute | Type             | Description  |
| --------- | ---------------- | ------------ |
| `key`     | `string`         | Object key   |
| `value`   | `TemplateString` | Object value |
