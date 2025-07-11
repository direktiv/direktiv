# Forms

- Forms will probably be the a feature that will challenge the Pages the most
- I assume that this will be the core feature that needs to be very flexible and extensible and work with
- prepare to haev the most exotic payloads that we can imagine

## Form primitves

I suggest we use the following primitives:

- input
  - type: "date" | "password" | "email" | "number" | "text" | "url"
  - data type: string or number (html will convert a value of a number to a string but I suggest we always cast this back to a number)
- textarea
- checkbox (multiple)
  - data type: boolean
- select
  - values: string[] ?s
  - data type: string
- file (I mostly inlcuded this because I wanna make sure that we already have a story on how we solve this)

Every primitive will have the following attributes:

- id (for referencing it)
- label
- description (can be empy)
- defaultValue
  - always a string
  - for

TBD:

- data types
- how to solve multiple checkboxes

# For simplicity, we don not add

- we don't add radio buttons, we can solve this data type with a select input

--

## Open Questions:

### Data types

When working with e.g. a mutation, the payload, headers and get parameters are from type `KeyValue`. Where the key is a `string` and the value is a `TemplateString`. The problem is, that `KeyValue` assumes that the key and the value are always strings. This maps very well to how to the data type of get parameters and request headers are shaped (they can only be strings), but the payload would need to be more flexible.

First, we should set some general boundaries for what type of payload we want to support. I would suggest that limit the playload to be from type `application/json` or `multipart/form-data` when a file upload is involved. This shoudl just be a convention that we don't even show the user (edge case: we should think about what we do if the user sets the header `Content-Type` by themselves, maybe we just overwrite it and/or show a warning if the user adds this key to a headers).

A `KeyValue`

- static: `{ "name": "John Doe" }`
- dynamic: `{ "name": "{{form.userForm.name}}" }`
- hybrid: `{ "action": "create-{{form.userForm.name}}", "age": 30 }`

File uploads: we could either go for application/json
