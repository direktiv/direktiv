import { TemplateStringType } from "../../../../schema/primitives/templateString";
import { replaceVariables } from "./utils";

type TemplateStringProps = {
  value: TemplateStringType;
};

export const TemplateString = ({ value }: TemplateStringProps) => (
  <span>{replaceVariables(value)}</span>
);
