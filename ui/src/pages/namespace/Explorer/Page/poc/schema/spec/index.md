# Page Document Spec

## Principles

- Page documents have no knowledge of the Direktiv Gateway. API requests are defined as generic fetch requests. However the UI can help the user pick a proper endpoint by suggesting routes that are available in the current environment. This principle makes a page document more portable and independent of the gateway's specification. This means if a user updates a gateway's endpoint, they also have to update all potential pages that use these endpoints.
- The specification does not restrict composition of elements that can result in invalid or unexpected behavior, similar to HTML when you combine tags inappropriately, like a button inside a button or a div inside a p tag. The user interface is more opinionated about what the user should be allowed to do. For example, it will not allow placing a modal inside another modal.

## Direktiv Page

| Attribute      | Type        | Description                      |
| -------------- | ----------- | -------------------------------- |
| `direktiv_api` | `"page/v1"` |                                  |
| `type`         | `"page"`    |                                  |
| `blocks`       | `Block[]`   | Entry point for all page content |

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

| Attribute | Type        | Description            |
| --------- | ----------- | ---------------------- |
| `type`    | `"columns"` |                        |
| `blocks`  | `Column[]`  | Array of column blocks |

### Column `Block`

| Attribute | Type       | Description                      |
| --------- | ---------- | -------------------------------- |
| `type`    | `"column"` |                                  |
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
| `data`    | `Block<Loop>`          | The loop block to iterate on                              |
| `actions` | `Block<Button>[]`      | List of actions that will be available in the last column |
| `columns` | `Block<TableColumn>[]` | List of table columns                                     |

### TableColumn `Block`

| Attribute | Type             | Description            |
| --------- | ---------------- | ---------------------- |
| `type`    | `"table-column"` |                        |
| `label`   | `TemplateString` | Headline of the column |
| `content` | `TemplateString` | Cell content           |

### Button `Block<trigger>`

| Attribute | Type             | Description                                                                                                        |
| --------- | ---------------- | ------------------------------------------------------------------------------------------------------------------ |
| `type`    | `"button"`       |                                                                                                                    |
| `label`   | `TemplateString` | Button label text                                                                                                  |
| `submit?` | `Mutation`       | Mutation that will be executed on click. This is optional as the button can also be used as a trigger for a dialog |

### Form `Block`

| Attribute  | Type             | Description                      |
| ---------- | ---------------- | -------------------------------- |
| `type`     | `"form"`         |                                  |
| `trigger`  | `Block<trigger>` | Trigger block to submit the form |
| `mutation` | `Mutation`       | Mutation executed on submission  |
| `blocks`   | `Block[]`        | Form content                     |

### Form Primitives

Form primitives are the basic input elements that collect user data within forms, such as text inputs, checkboxes, and dropdowns. Form primitives must be placed inside a form block. They can be nested at any depth within other blocks, but one form block must exist somewhere up the tree. All form primitives share these common fields:

| Attribute     | Type             | Description                                           |
| ------------- | ---------------- | ----------------------------------------------------- |
| `id`          | `Id`             | Unique identifier for the field                       |
| `label`       | `TemplateString` | Field label text                                      |
| `description` | `TemplateString` | Field description. Can be empty except for checkboxes |
| `optional`    | `boolean`        | Whether the field is optional                         |

#### String Input `Block`

| Attribute      | Type                                       | Description   |
| -------------- | ------------------------------------------ | ------------- |
| `type`         | `"form-string-input"`                      |               |
| `variant`      | `"text"`, `"password"`, `"email"`, `"url"` | Input type    |
| `defaultValue` | `TemplateString`                           | Default value |

#### Number Input `Block`

| Attribute      | Type                             | Description                                                                                      |
| -------------- | -------------------------------- | ------------------------------------------------------------------------------------------------ |
| `type`         | `"form-number-input"`            |                                                                                                  |
| `defaultValue` | `Number`, `VariablePath<number>` | Default value. Either a static number or a variable path to a variable that resolves to a number |

#### Date Input `Block`

| Attribute      | Type                | Description                                                                                                                    |
| -------------- | ------------------- | ------------------------------------------------------------------------------------------------------------------------------ |
| `type`         | `"form-date-input"` |                                                                                                                                |
| `defaultValue` | `string`            | Default date value. Can be every value that JavaScript's `new Date()` accepts. E.g. `2025-12-24T00:00:00.000Z` or `2025-12-24` |

#### Select `Block`

| Attribute      | Type                                   | Description                                                                                      |
| -------------- | -------------------------------------- | ------------------------------------------------------------------------------------------------ |
| `type`         | `"form-select"`                        |                                                                                                  |
| `values`       | `Array`, `VariablePath<Array<string>>` | Available options. Either a static array or a variable path that resolves to an array of strings |
| `defaultValue` | `TemplateString`                       | Default selected value                                                                           |

#### Textarea `Block`

| Attribute      | Type              | Description        |
| -------------- | ----------------- | ------------------ |
| `type`         | `"form-textarea"` |                    |
| `defaultValue` | `TemplateString`  | Default text value |

#### Checkbox `Block`

| Attribute      | Type                               | Description                                                                          |
| -------------- | ---------------------------------- | ------------------------------------------------------------------------------------ |
| `type`         | `"form-checkbox"`                  |                                                                                      |
| `description`  | `TemplateString`                   | Required description text                                                            |
| `defaultValue` | `Boolean`, `VariablePath<boolean>` | Default value. Either a static boolean or a variable path that resolves to a boolean |

### QueryProvider `Block`

A QueryProvider is a block that is responsible for fetching data from one or multiple APIs and providing the result to its child blocks. It will display a loading indicator until all data is fetched successfully. Every child block can access the data in fields of type `TemplateString` by using the `id` of the corresponding `Query` inside the `QueryProvider` as a reference.

| Attribute | Type               | Description                      |
| --------- | ------------------ | -------------------------------- |
| `type`    | `"query-provider"` |                                  |
| `queries` | `Query[]`          | Queries to execute               |
| `blocks`  | `Block[]`          | Children with access to the data |

### Loop `Block`

A Block that allows to iterate over an array variable. It renders its blocks for each item.

| Attribute  | Type                          | Description                                                                       |
| ---------- | ----------------------------- | --------------------------------------------------------------------------------- |
| `type`     | `"loop"`                      |                                                                                   |
| `id`       | `Id`                          | Unique identifier                                                                 |
| `data`     | `VariablePath<Array<Object>>` | Variable path to an array of objects to loop over                                 |
| `pageSize` | `number`                      | Number of items per page                                                          |
| `blocks`   | `Block[]`                     | Child blocks rendered per item. Does not exist when used as a context for a table |

### Image `Block`

| Attribute | Type             | Description            |
| --------- | ---------------- | ---------------------- |
| `type`    | `"image"`        |                        |
| `src`     | `TemplateString` | Image source URL       |
| `width`   | `number`         | Image width in pixels  |
| `height`  | `number`         | Image height in pixels |

**Example**

When adding a `loop` block with the id `pokemon` and a variable `query.pokemonList.data` (points to a `query` with the id `pokemon-list` and references the `data` array returned from the API request), every block will have access to the variable `loop.pokemon` to reference its corresponding array item.

# Procedures

Procedures are types that can be used in various blocks that will handle API requests

### `Mutation`

A mutation is an API request that modifies data on the server

| Attribute         | Type                 | Description               |
| ----------------- | -------------------- | ------------------------- |
| `method`          | `MutationMethod`     | HTTP method               |
| `url`             | `TemplateString`     | URL to the API endpoint   |
| `queryParams?`    | `KeyValue[]`         | Optional query parameters |
| `requestHeaders?` | `KeyValue[]`         | Optional request headers  |
| `requestBody?`    | `ExtendedKeyValue[]` | Optional request body     |

### `MutationMethod`

`"POST"`, `"PUT"`, `"PATCH"`, `"DELETE"`

## `Query`

A query is a API `GET`-request that reads data from the server

| Attribute      | Type             | Description                            |
| -------------- | ---------------- | -------------------------------------- |
| `id`           | `Id`             | Unique ID used to reference query data |
| `url`          | `TemplateString` | URL to the API endpoint                |
| `queryParams?` | `KeyValue[]`     | Optional query parameters              |

# Primitives

## `Variable`

A Variable is a string that references dynamic data from various sources within the page. Variables follow the structure `namespace.id.pointer` and are resolved at runtime to access contextual data.

### Structure

- **namespace**: The source of the data (e.g., `query`, `loop`, `this`)
- **id**: The identifier of the specific block or context
- **pointer**: The path to the specific data within the source (not available in the `this` namespace)

### Available Namespaces

- **`query`**: References data from Query blocks. The `id` is the query's unique identifier, and the `pointer` navigates the JSON response.
- **`loop`**: References data from Loop blocks. The `id` is the loop's unique identifier, and the `pointer` accesses the current item in the iteration.
- **`this`**: References local variables within the current context, such as form submission data in a form block. Currently, `form` is the only block type that supports the `this` namespace. In this namespace, the `id` specifies the form primitive and `pointer` is not supported as the primitive already holds the value and no further pointer is needed.

### Variable Scoping

Variables are scoped based on their namespace:

- **`query`** and **`loop`** variables are available from the block where they are defined and propagate downward through the component tree to child blocks.
- **`this`** variables are only available within the block itself (e.g., form data is accessible only inside the form block).

**Examples**

- `query.user.data.name`

  - `query` (namespace): References data from a query within a QueryProvider block
  - `user` (id): The specific query with id "user"
  - `data.name` (pointer): Navigates to the `name` field in the `data` object of the JSON response

- `loop.items.data.title`

  - `loop` (namespace): References data from a Loop block
  - `items` (id): The specific loop block with id "items"
  - `title` (pointer): Accesses the `title` field of the current item being iterated over

- `this.username`
  - `this` (namespace): References local variables within the current context
  - `username` (id): Searches for the form primitive with the id "username" and uses that value

## `TemplateString`

A template string is a string that can have `Variable` placeholders that will be filled with dynamic data. Variables in template strings will always be stringified if possible. If a variable cannot be stringified, it will throw an error.

**Example**

`Edit {{query.user.data.name}}` will be resolved to `Edit John Doe` if the query with id "user" returns `{ data: { name: "John Doe" } }`.

## `Id`

An Id is a string that identifies a block and must be unique among all blocks of the same type. IDs are part of a variable and cannot contain dots (.) as they are used as separators for variables.

## `KeyValue`

| Attribute | Type             | Description  |
| --------- | ---------------- | ------------ |
| `key`     | `string`         | Object key   |
| `value`   | `TemplateString` | Object value |

## `ExtendedKeyValue`

An extended key-value pair that supports multiple data types for the value, including strings, variables, booleans, and numbers.

| Attribute | Type                                          | Description  |
| --------- | --------------------------------------------- | ------------ |
| `key`     | `string`                                      | Object key   |
| `value`   | `Number`, `String`, `Boolean`, `VariablePath` | Object value |

## `Number`

| Attribute | Type       | Description  |
| --------- | ---------- | ------------ |
| `key`     | `"number"` |              |
| `value`   | `number`   | Number value |

## `String`

| Attribute | Type             | Description  |
| --------- | ---------------- | ------------ |
| `key`     | `"string"`       |              |
| `value`   | `TemplateString` | Object value |

## `Boolean`

| Attribute | Type        | Description          |
| --------- | ----------- | -------------------- |
| `key`     | `"boolean"` |                      |
| `value`   | `boolean`   | Either true or false |

## `Array`

| Attribute | Type      | Description         |
| --------- | --------- | ------------------- |
| `key`     | `"array"` |                     |
| `value`   | `array`   | An array of strings |

## `VariablePath`

| Attribute | Type         | Description          |
| --------- | ------------ | -------------------- |
| `key`     | `"variable"` |                      |
| `value`   | `Variable`   | a path to a variable |
