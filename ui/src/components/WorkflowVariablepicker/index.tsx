import { Fragment, useState } from "react";
import {
  Variablepicker,
  VariablepickerError,
  VariablepickerHeading,
  VariablepickerItem,
  VariablepickerMessage,
  VariablepickerSeparator,
} from "~/design/VariablePicker";

import { ButtonBar } from "~/design/ButtonBar";
import Input from "~/design/Input";
import { VarSchema } from "~/api/variables/schema";
import { useNamespace } from "~/util/store/namespace";
import { useTranslation } from "react-i18next";
import { useWorkflowVariables } from "~/api/tree/query/variables";
import { z } from "zod";

const variableWithoutChecksumAndSize = VarSchema.omit({
  checksum: true,
  size: true,
});

type variableType = z.infer<typeof variableWithoutChecksumAndSize>;

const WorkflowVariablePicker = ({
  namespace: givenNamespace,
  workflow,
  defaultVariable,
  onChange,
}: {
  namespace?: string;
  workflow: string;
  defaultVariable?: variableType;
  onChange: (variable: variableType | undefined) => void;
  //onChange?: (event: React.FormEvent) => void;
}) => {
  const { t } = useTranslation();

  //   const defaultNamespace = useNamespace();

  const namespace = givenNamespace ? givenNamespace : undefined;
  const path = workflow;

  console.log("ns  " + namespace);
  console.log("p " + path);

  //   if (!namespace) {
  //     throw new Error("namespace is undefined");
  //   }

  //const path = workflow ? workflow : "/";

  const { data, isError } = useWorkflowVariables({ path, namespace });

  if (isError) {
    throw new Error("path not found");
  }

  const variableList = data?.variables.results
    ? data.variables.results
    : undefined;

  if (!variableList) {
    throw new Error("namespace is undefined");
  }

  const [inputValue, setInput] = useState(
    defaultVariable ? defaultVariable.name : ""
  );

  const [variable, setVariable] = useState(
    defaultVariable ? defaultVariable : undefined
  );

  const [index, setIndex] = useState(0);

  const emptyVariable: variableType = {
    name: "",
    createdAt: "",
    updatedAt: "",
    mimeType: "",
  };

  const pathNotFound = isError;

  const handleIt = (value: string) => {
    setInput(value);
    setIndex(-1);
    emptyVariable.name = value;
    setVariable(emptyVariable);
    onChange(emptyVariable);
  };

  const handleChanges = (index: number) => {
    setIndex(index);

    if (
      variableList != undefined &&
      variableList[index] != undefined &&
      variableList?.[index]?.name != undefined
    ) {
      const newVar = variableList?.[index];
      const newVarName = variableList?.[index]?.name;

      newVarName === undefined ? setInput(inputValue) : setInput(newVarName);
      newVar === undefined ? onChange(undefined) : onChange(newVar);
      variableList?.[index] != undefined
        ? setVariable(variableList[index])
        : setVariable(emptyVariable);
    }
  };

  return (
    <>
      <ButtonBar>
        {pathNotFound ? (
          <VariablepickerError
            buttonText={t("components.workflowVariablepicker.buttonText")}
          >
            <VariablepickerHeading>
              {t("components.workflowVariablepicker.title", { path })}
            </VariablepickerHeading>
            <VariablepickerSeparator />

            <VariablepickerMessage>
              {t("components.workflowVariablepicker.error.title", { path })}
            </VariablepickerMessage>
            <VariablepickerSeparator />
          </VariablepickerError>
        ) : (
          <>
            {!variableList || variableList.length == 0 ? (
              <VariablepickerError
                buttonText={t("components.workflowVariablepicker.buttonText")}
              >
                <VariablepickerHeading>
                  {t("components.workflowVariablepicker.title", { path })}
                </VariablepickerHeading>
                <VariablepickerSeparator />

                <VariablepickerMessage>
                  {t("components.workflowVariablepicker.emptyDirectory.title", {
                    path,
                  })}
                </VariablepickerMessage>
                <VariablepickerSeparator />
              </VariablepickerError>
            ) : (
              <Variablepicker
                buttonText={t("components.workflowVariablepicker.buttonText")}
                onChange={onChange}
                onValueChange={(index) => handleChanges(index)}
              >
                <VariablepickerHeading>
                  {t("components.workflowVariablepicker.title", { path })}
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
          placeholder={t("components.workflowVariablepicker.placeholder")}
          value={inputValue}
          onChange={(e) => {
            handleIt(e.target.value);
          }}
        />
      </ButtonBar>
    </>
  );
};

export default WorkflowVariablePicker;
