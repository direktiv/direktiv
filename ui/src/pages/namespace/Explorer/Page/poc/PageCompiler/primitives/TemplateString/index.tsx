import { TemplateStringType } from "../../../schema/primitives/templateString";
import { Variable } from "../Variable";
import { processTemplateString } from "./utils";

type TemplateStringProps = {
  value: TemplateStringType;
};

export const TemplateString = ({ value }: TemplateStringProps) => {
  const templateFragments = processTemplateString(value, (match, index) => (
    <Variable key={index} value={match} />
  ));

  return templateFragments;
};
