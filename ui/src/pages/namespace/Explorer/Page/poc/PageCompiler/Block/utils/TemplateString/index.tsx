import { TemplateStringType } from "../../../../schema/primitives/templateString";
import { Variable } from "./Variable";
import { variablePattern } from "./Variable/utils";

type TemplateStringProps = {
  value: TemplateStringType;
};

export const TemplateString = ({ value }: TemplateStringProps) => {
  const templateFragments = value.split(variablePattern);
  return (
    <>
      {templateFragments.map((fragment, index) => {
        const isVariable = index % 2 === 1;

        if (isVariable) {
          return <Variable key={index} value={fragment} />;
        }

        return fragment;
      })}
    </>
  );
};
