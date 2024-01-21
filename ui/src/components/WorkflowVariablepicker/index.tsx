import { FC, Fragment, PropsWithChildren, useState } from "react";
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
import { analyzePath } from "~/util/router/utils";
import { useNamespace } from "~/util/store/namespace";
import { useNodeContent } from "~/api/tree/query/node";
import { useTranslation } from "react-i18next";
import { useWorkflowVariables } from "~/api/tree/query/variables";
import { z } from "zod";

const variableWithoutChecksum = VarSchema.omit({ checksum: true });

type variableType = z.infer<typeof variableWithoutChecksum>;
const convertFileToPath = (string?: string) =>
  analyzePath(string).parent?.absolute ?? "/";

const defaultPath = "/";

const WorkflowVariables: FC<PropsWithChildren> = ({ children }) => {
  const [path, setPath] = useState(convertFileToPath(defaultPath));

  const namespace = useNamespace() ?? "/";
  const { data, isError } = useNodeContent({
    path,
    namespace,
  });

  const results = data?.children?.results ?? [];
  const workflows = data?.children?.results.filter(
    (element) => element.type === "workflow"
  );

  return (
    <Fragment>
      {workflows &&
        workflows.map((file) => (
          <>
            <Fragment key={file.name}>{file.name}</Fragment>
            {children}
            <br />
          </>
        ))}
    </Fragment>
  );
};

const WorkflowVariablePicker = ({
  namespace: givenNamespace,
  workflow,
  defaultVariable,
  onChange,
}: {
  namespace?: string;
  workflow: string;
  defaultVariable?: variableType;
  defaultPath?: string;
  onChange: (variable: variableType | undefined) => void;
}) => {
  const { t } = useTranslation();

  const defaultNamespace = useNamespace();

  const namespace = givenNamespace ? givenNamespace : defaultNamespace;
  if (!namespace) {
    throw new Error("namespace is undefined");
  }

  // BEFORE for Testing
  // const path = "workflow.yaml";

  const path = workflow;

  const { data, isError } = useWorkflowVariables({ path });

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
              <VariablepickerError buttonText={buttonText}>
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
                buttonText={buttonText}
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
            setInput(e.target.value);
            setIndex(-1);
          }}
        />
      </ButtonBar>
      <Button variant="outline">{variable?.createdAt}</Button>
      <Button variant="outline">{variable?.mimeType}</Button>
      <br></br>
      <Button>{inputValue}</Button>
      <br></br>
      <br></br>
      <Button>{JSON.stringify(variable)}</Button>
      <br></br>
      <Button>{index}</Button>
      <br></br>
    </>
  );
};

export default WorkflowVariablePicker;
