import { parseVariable, validateVariable } from "./utils";

import { Error } from "./Error";
import { QueryVariable } from "./Query";
import { VariableType } from "../../../../../schema/primitives/variable";

type VariablesProps = {
  value: VariableType;
};

export const Variable = ({ value }: VariablesProps) => {
  const [variable, error] = validateVariable(parseVariable(value));

  if (error) {
    return <Error value={value}>{error}</Error>;
  }

  const { id, namespace, pointer } = variable;
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
