import { TemplateStringType } from "../../../schema/primitives/templateString";

type TemplateStringProps = {
  value: TemplateStringType;
};

export const TemplateString = ({ value }: TemplateStringProps) => (
  <span>{value}</span>
);
