import {
  VariableType,
  supportedVariableNamespaces,
} from "../../../../../schema/primitives/variable";

import { Error } from "./Error";
import { QueryVariable } from "./Query";
import { parseVariable } from "./utils";

type VariablesProps = {
  value: VariableType;
};

export const Variable = ({ value }: VariablesProps) => {
  const { namespace, id, pointer } = parseVariable(value);

  if (!namespace)
    return (
      <Error value={value}>
        Could not find a matching variable namespace for <code>{value}</code>.
        Make sure the variable starts with on of the following namespaces:
        <ul className="list-disc pl-5 space-y-1">
          {supportedVariableNamespaces.map((namespace) => (
            <li key={namespace}>
              <code>{namespace}</code>
            </li>
          ))}
        </ul>
      </Error>
    );

  if (!id)
    return (
      <Error value={value}>
        Could not find a any id. Please add an id to your variable to point to
        the corresponding <code>{namespace}</code>.
      </Error>
    );

  if (!pointer)
    return (
      <Error value={value}>
        Could not find any variable pointer. Please add a path to a variable
        that is available in the <code>{namespace}</code> with the id{" "}
        <code>{id}</code>.
      </Error>
    );

  switch (namespace) {
    case "query":
      return <QueryVariable id={id} pointer={pointer} />;
      break;
    default:
      return (
        <Error value={value}>
          There is no implementation for <code>{namespace}</code> yet.
        </Error>
      );
      break;
  }
};
