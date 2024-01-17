import { Fragment, useState } from "react";
import {
  Variablepicker,
  VariablepickerError,
  VariablepickerHeading,
  VariablepickerItem,
  VariablepickerMessage,
  VariablepickerSeparator,
} from "~/design/VariablePicker";

import Button from "~/design/Button";
import { ButtonBar } from "~/design/ButtonBar";
import Input from "~/design/Input";
import { VarSchema } from "~/api/variables/schema";
import { useNamespace } from "~/util/store/namespace";
import { useTranslation } from "react-i18next";
import { useVars } from "~/api/variables/query/useVariables";
import { z } from "zod";

type variableType = z.infer<typeof VarSchema>;

const NamespaceVariablePicker = ({
  namespace,
  defaultVariable,
  onChange,
}: {
  namespace?: string;
  defaultVariable?: variableType;
  onChange: (variable: variableType | undefined) => void;
}) => {
  const { t } = useTranslation();

  const path = useNamespace();

  const { data, isError } = useVars({ namespace });

  const variableList = data?.variables.results
    ? data.variables.results
    : undefined;

  const [inputValue, setInput] = useState(
    defaultVariable ? defaultVariable.name : ""
  );

  const [index, setIndex] = useState(0);

  const [variable, setVariable] = useState(
    defaultVariable ? defaultVariable : undefined
  );

  const buttonText = "Browse Variables";

  const pathNotFound = isError;
  const emptyVariableList = !isError && !variableList?.length;

  const varname = variable ? variable.name : "test";

  console.log("variable " + varname);
  console.log("index " + index);
  console.log("inputValue " + inputValue);

  const handleChanges = (index: number) => {
    setIndex(index);
    if (
      variableList != undefined &&
      variableList[index] != undefined &&
      variableList?.[index]?.name != undefined
    ) {
      const newVar = variableList?.[index];
      const newVarName = variableList?.[index]?.name;

      console.log("newVar " + newVar);
      console.log("newVarName " + newVarName);

      const test = variableList?.[index];
      console.log("index is " + test);

      newVarName === undefined ? setInput(inputValue) : setInput(newVarName);
      newVar === undefined ? onChange(undefined) : onChange(newVar);
      variableList?.[index] != undefined
        ? setVariable(variableList[index])
        : undefined;
    }
  };

  return (
    <>
      <ButtonBar>
        {pathNotFound ? (
          <VariablepickerError buttonText={buttonText}>
            <VariablepickerHeading>
              {t("components.namespaceVariablepicker.title", { path })}
            </VariablepickerHeading>
            <VariablepickerSeparator />

            <VariablepickerMessage>
              {t("components.namespaceVariablepicker.error.title", { path })}
            </VariablepickerMessage>
            <VariablepickerSeparator />
          </VariablepickerError>
        ) : (
          <>
            {!variableList || variableList.length == 0 ? (
              <VariablepickerError buttonText={buttonText}>
                <VariablepickerHeading>
                  {t("components.namespaceVariablepicker.title", { path })}
                </VariablepickerHeading>
                <VariablepickerSeparator />

                <VariablepickerMessage>
                  {t(
                    "components.namespaceVariablepicker.emptyDirectory.title",
                    { path }
                  )}
                </VariablepickerMessage>
                <VariablepickerSeparator />
              </VariablepickerError>
            ) : (
              <Variablepicker
                buttonText={buttonText}
                onChange={onChange}
                onValueChange={(index) => handleChanges(index)}
              >
                <VariablepickerHeading>
                  {t("components.namespaceVariablepicker.title", { path })}
                </VariablepickerHeading>
                <VariablepickerSeparator />
                {variableList.map((variable, index) => (
                  <Fragment key={index}>
                    <VariablepickerItem value={index}>
                      {variable.name}
                    </VariablepickerItem>
                    <VariablepickerSeparator />
                  </Fragment>
                ))}
              </Variablepicker>
            )}
          </>
        )}
        <Input
          placeholder={t("components.namespaceVariablepicker.placeholder")}
          value={inputValue}
          onChange={(e) => {
            setInput(e.target.value);
            setIndex(-1);
          }}
        />
      </ButtonBar>
      <Button variant="outline">{variable?.createdAt}</Button>
    </>
  );
};

export default NamespaceVariablePicker;
