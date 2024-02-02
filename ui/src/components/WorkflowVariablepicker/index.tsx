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
import { VariablePickerError } from "./VariablePickerError";
import { useNamespace } from "~/util/store/namespace";
import { useTranslation } from "react-i18next";
import { useWorkflowVariables } from "~/api/tree/query/variables";

type WorkflowVariablePickerProps = {
  namespace?: string;
  workflowPath: string;
  defaultVariable?: string;
  onChange: (name: string, mimeType?: string) => void;
};

const WorkflowVariablePicker = ({
  namespace: givenNamespace,
  workflowPath,
  defaultVariable,
  onChange,
}: WorkflowVariablePickerProps) => {
  const { t } = useTranslation();

  const [inputValue, setInput] = useState(defaultVariable ?? "");

  const defaultNamespace = useNamespace();
  const namespace = givenNamespace ?? defaultNamespace;

  const { data, isError: pathNotFound } = useWorkflowVariables({
    path: workflowPath,
    namespace,
  });

  const variables = data?.variables.results ?? [];
  const noVarsInWorkflow = variables.length === 0;

  const setNewVariable = (name: string) => {
    onChange(name);
    setInput(name);
  };

  const setExistingVariable = (name: string) => {
    const foundVariable = variables.find((element) => element.name === name);
    if (foundVariable) {
      onChange(foundVariable?.name, foundVariable?.mimeType);
      setInput(name);
    }
  };

  const getErrorComponent = () => {
    if (!workflowPath) {
      return (
        <VariablePickerError>
          {t("components.workflowVariablepicker.unselected.title")}
        </VariablePickerError>
      );
    }

    if (pathNotFound) {
      return (
        <VariablepickerError
          buttonText={t("components.workflowVariablepicker.buttonText")}
        >
          <VariablepickerHeading>
            {t("components.workflowVariablepicker.title", {
              path: workflowPath,
            })}
          </VariablepickerHeading>
          <VariablepickerSeparator />
          <VariablepickerMessage>
            {t("components.workflowVariablepicker.error.title", {
              path: workflowPath,
            })}
          </VariablepickerMessage>
          <VariablepickerSeparator />
        </VariablepickerError>
      );
    }

    if (noVarsInWorkflow) {
      return (
        <VariablepickerError
          buttonText={t("components.workflowVariablepicker.buttonText")}
        >
          <VariablepickerHeading>
            {t("components.workflowVariablepicker.title", {
              path: workflowPath,
            })}
          </VariablepickerHeading>
          <VariablepickerSeparator />
          <VariablepickerMessage>
            {t("components.workflowVariablepicker.emptyDirectory.title", {
              path: workflowPath,
            })}
          </VariablepickerMessage>
          <VariablepickerSeparator />
        </VariablepickerError>
      );
    }
    return null;
  };

  const errorComponent = getErrorComponent();

  return (
    <ButtonBar>
      {errorComponent ?? (
        <Variablepicker
          buttonText={t("components.workflowVariablepicker.buttonText")}
          value={inputValue}
          onValueChange={(value) => {
            setExistingVariable(value);
          }}
        >
          <VariablepickerHeading>
            {t("components.workflowVariablepicker.title", {
              path: workflowPath,
            })}
          </VariablepickerHeading>
          <VariablepickerSeparator />
          {variables.map((variable, index) => (
            <Fragment key={index}>
              <VariablepickerItem value={variable.name}>
                {variable.name}
              </VariablepickerItem>
              <VariablepickerSeparator />
            </Fragment>
          ))}
        </Variablepicker>
      )}
      <Input
        placeholder={t("components.workflowVariablepicker.placeholder")}
        value={inputValue}
        onChange={(e) => {
          setNewVariable(e.target.value);
        }}
      />
    </ButtonBar>
  );
};

export default WorkflowVariablePicker;
