# Forms

- Forms will probably be the a feature that will challenge the Pages the most
- I assume that this will be the core feature that needs to be very flexible
- Here is a small list of features that I think we need to support:

  - make static forms like
    - text inputs: name, email, etc
    - select from a static list: country, language, etc
  - make dynamic forms likey

    - select from a list of values from a query result

## Form primitves

Form primitives should all be pretty straight forward, they are all just blocks that adds a variable to the form that they are in. Every form primitive returns a specific data type (text area -> string, checkbox -> boolean, etc). It does not matter how deep they are nested.

The actual usage of the variables will be done in the form component itself. This is were the actual complexity needs to be solved.

### Form primitves

I suggest we use the following primitives:

- input
  - type:
    - text (string)
    - date (string)
    - password (string)
    - email (string)
    - url (string)
    - number (number)
- textarea (string)
- checkbox (boolean)
- select (string)
  - values: needs a list of strings as a default value input
    - could either be static or dynamic
- file
  - I mostly included this because I wanna make sure that we already have a story on how we solve thi. We will definitely need to supoprt this at some point
  - which type should it be?
    - base64 encoded string would be the easiest to implement because it is just a string and therefore JSON compatible
    - binary would be the more common format (multipart/form-data)

Every primitive will have the following attributes in common:

- id (for referencing it in the form block), technically not needed as an identifier when we don't allow nesting forms, but we probably need to support some kind of nesting as nesting block forms would not nessesarily be also nested form tags in the dom (which would be invalid html). Example: the user could have a form on a root level of the page and then a have a modal openeing a form
- label
- description (can be empy)
- defaultValue (always the same data type as the output data)
  - can be dynamic or static

Example of some form primitives in JSON:

## Extending the KeyValue Schema

This is how KeyValue looks right now:

```
 queryParams: [
  {
    key: "query",
    value: "string",
  },
],
```

This schema only supports strings for key and values. Values here can even use variable strings make use of existing variables from type string. This works fine in mutations/queries when defining headers and get parameters, as no other data type is supported.

are from type `KeyValue`. Where the key is a `string` and the value is a `TemplateString`. The problem is, that `KeyValue` assumes that the key and the value are always strings. This maps very well to how the data type of get parameters and request headers are shaped (they can only be strings), but the payload of a mutation would need to be more flexible and support JSON at least (the values need to support: string, number, boolean, null, array, object).

## Form primitves

Every primitive will have the following attributes:

TBD:

- data types
- how to solve multiple checkboxes (user needs to compose an array)
- validation

# For simplicity, we don not add

- we don't add radio buttons, we can solve this data type with a select input

# Future considerations

- It will probably be a requirement to having multi step forms (like a multi step wizard) and I think this would be a solid foundation for that

## Open Questions:

First, we should set some general boundaries for what type of payload we want to support. I would suggest that limit the playload to be from type `application/json` or `multipart/form-data` when a file upload is involved. This shoudl just be a convention that we don't even show the user (edge case: we should think about what we do if the user sets the header `Content-Type` by themselves, maybe we just overwrite it and/or show a warning if the user adds this key to a headers).

A `KeyValue`

- static: `{ "name": "John Doe" }`
- dynamic: `{ "name": "{{form.userForm.name}}" }`
- hybrid: `{ "action": "create-{{form.userForm.name}}", "age": 30 }`

File uploads: we could either go for application/json and allow
