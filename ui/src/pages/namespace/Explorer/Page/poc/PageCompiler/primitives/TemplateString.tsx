import { TemplateStringType } from "../../schema/primitives/templateString";
import { Variable } from "./Variable";
import { parseTemplateString } from "./Variable/utils";

type TemplateStringProps = {
  value: TemplateStringType;
};

export const TemplateString = ({ value }: TemplateStringProps) => {
  const templateFragments = parseTemplateString(value, (match, index) => (
    <Variable key={index} value={match} />
  ));

  return templateFragments;
};
