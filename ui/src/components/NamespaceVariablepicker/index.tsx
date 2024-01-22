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
  namespace: givenNamespace,
  defaultVariable,
  onChange,
}: {
  namespace?: string;
  defaultVariable?: variableType;
  //onChange: (value: number) => void;
  onChange: (variable: variableType | undefined) => void;
  //onChange?: (event: React.FormEvent) => void;
}) => {
  const { t } = useTranslation();
  const defaultNamespace = useNamespace();
  defaultNamespace;
  const namespace = givenNamespace ? givenNamespace : defaultNamespace;

  console.log("namespace " + namespace);
  const { data, isError } = useVars({ namespace });
  const path = namespace;
  console.log("path " + path);
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

  const emptyVariable: variableType = {
    name: "",
    createdAt: "123",
    updatedAt: "",
    mimeType: "",
    checksum: "",
    size: "",
  };

  const pathNotFound = isError;
  const emptyVariableList = !isError && !variableList?.length;

  const varname = variable ? variable.name : "test";

  const handleChanges = (index: number) => {
    setIndex(index);
    console.log("here");
    if (
      variableList != undefined &&
      variableList[index] != undefined &&
      variableList?.[index]?.name != undefined
    ) {
      const newVariable = variableList[index];

      if (newVariable === undefined) {
        throw new Error("Variable is undefined");
      }
      console.log("There");
      // const newVarName = variableList?.[index]?.name;

      // const test = variableList?.[index];

      setInput(newVariable.name);
      setVariable(newVariable);

      // newVarName === undefined ? setInput(inputValue) : setInput(newVarName);
      // newVar === undefined ? onChange(undefined) : onChange(newVar);
      // variableList?.[index] != undefined
      //   ? setVariable(variableList[index])
      //   : undefined;
    }
  };

  const handleIt = (value: string) => {
    setInput(value);
    setIndex(-1);
    emptyVariable.name = value;
    setVariable(emptyVariable);
    onChange(emptyVariable);
  };

  return (
    <>
      <ButtonBar>
        {pathNotFound ? (
          <VariablepickerError
            buttonText={t("components.namespaceVariablepicker.buttonText")}
          >
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
              <VariablepickerError
                buttonText={t("components.namespaceVariablepicker.buttonText")}
              >
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
                buttonText={t("components.namespaceVariablepicker.buttonText")}
                onChange={(e) => onChange(variable)}
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
            handleIt(e.target.value);
          }}
        />
      </ButtonBar>
      <Button variant="outline">{variable?.createdAt}</Button>
    </>
  );
};

export default NamespaceVariablePicker;
