import {
  ArrayFieldTemplateItemType,
  ArrayFieldTemplateProps,
  BaseInputTemplateProps,
  DescriptionFieldProps,
  ErrorListProps,
  RJSFSchema,
  RJSFValidationError,
  RegistryWidgetsType,
  TitleFieldProps,
  WidgetProps,
} from "@rjsf/utils";
import {
  ChevronDownIcon,
  ChevronUpIcon,
  MinusIcon,
  PlusIcon,
} from "lucide-react";
import React, { ComponentProps, useMemo } from "react";
import {
  Select,
  SelectContent,
  SelectGroup,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "../Select";

import Alert from "~/design/Alert";
import Button from "../Button";
import { Checkbox } from "../Checkbox";
import Form from "@rjsf/core";
import Input from "../Input";
import { twMergeClsx } from "~/util/helpers";
import validator from "@rjsf/validator-ajv8";

const CustomSelectWidget: React.FC<WidgetProps> = (props) => (
  <div className="my-4">
    <Select onValueChange={props.onChange} value={`${props.value ?? ""}`}>
      <SelectTrigger id={props.id}>
        <SelectValue>{props.value ?? `Select ${props.label}`}</SelectValue>
      </SelectTrigger>
      <SelectContent>
        <SelectGroup>
          {props.options.enumOptions?.map((op) => (
            <SelectItem key={`select-${op.value}`} value={`${op.value ?? ""}`}>
              {op.label}
            </SelectItem>
          ))}
        </SelectGroup>
      </SelectContent>
    </Select>
  </div>
);

const SubmitButton = () => null;

const ArrayFieldTemplateItem = (props: ArrayFieldTemplateItemType) => (
  <div className="flex flex-row items-end gap-4">
    <div className="grow">{props.children}</div>
    <div className="mb-2 flex w-min flex-row gap-2">
      <Button
        disabled={!props.hasMoveDown}
        variant="outline"
        onClick={(e) => {
          props.onReorderClick(props.index, props.index + 1)(e);
        }}
        data-testid={`json-schema-form-down-button-${props.index}`}
        icon
        type="button"
      >
        <ChevronDownIcon />
      </Button>
      <Button
        disabled={!props.hasMoveUp}
        variant="outline"
        onClick={(e) => {
          props.onReorderClick(props.index, props.index - 1)(e);
        }}
        data-testid={`json-schema-form-up-button-${props.index}`}
        icon
        type="button"
      >
        <ChevronUpIcon />
      </Button>
      <Button
        variant="outline"
        onClick={(e) => {
          props.onDropIndexClick(props.index)(e);
        }}
        data-testid={`json-schema-form-remove-button-${props.index}`}
        icon
        type="button"
      >
        <MinusIcon />
      </Button>
    </div>
  </div>
);

const ArrayFieldTemplate = (props: ArrayFieldTemplateProps) => (
  <div className="my-4">
    <div className="my-2 flex items-center">
      <div className="grow">{props.title}</div>
    </div>
    {props.items.map((element) => (
      <ArrayFieldTemplateItem {...element} key={`array-item-${element.key}`} />
    ))}
    <Button
      onClick={props.onAddClick}
      icon
      disabled={!props.canAdd}
      variant="outline"
      data-testid="json-schema-form-add-button"
      type="button"
      block
      className="mt-3"
    >
      <PlusIcon />
    </Button>
  </div>
);

const BaseInputTemplate = (props: BaseInputTemplateProps) => {
  const type = useMemo(() => {
    if (props.schema.type === "integer") {
      return "number";
    } else if (props.type === "file") {
      return "file";
    } else {
      return undefined;
    }
  }, [props.schema.type, props.type]);

  return (
    <Input
      defaultValue={props.value}
      className="mb-2 mt-1 w-full"
      type={type}
      required={props.required}
      id={props.id}
      onChange={(e) => {
        if (type === "file") {
          if (e.target.files && e.target.files.length > 0) {
            const reader = new FileReader();
            reader.onloadend = () => {
              props.onChange(reader.result);
            };
            reader.readAsDataURL(e.target.files[0] as Blob);
          } else {
            props.onChange(e.target.files);
          }
        } else {
          props.onChange(e.target.value);
        }
      }}
    />
  );
};

const TitleFieldTemplate = (props: TitleFieldProps) => {
  const { id, required, title } = props;
  return (
    <header id={id} className="mb-4 font-semibold">
      {title}
      {required && <mark>*</mark>}
    </header>
  );
};
const DescriptionFieldTemplate = (props: DescriptionFieldProps) => {
  const { description } = props;
  return (
    <div className="mb-2 text-gray-10 dark:text-gray-dark-10">
      {description}
    </div>
  );
};

const ErrorListTemplate = (props: ErrorListProps) => {
  const { errors } = props;
  return (
    <Alert
      variant="error"
      data-testid="jsonschema-form-error"
      className="mb-2"
      {...props}
    >
      <ul>
        {errors.map((error: RJSFValidationError, i: number) => (
          <li key={i}>{`${error.stack}`}</li>
        ))}
      </ul>
    </Alert>
  );
};

const CustomCheckbox = (props: WidgetProps) => (
  <div className="flex space-x-2 p-2 ">
    <Checkbox
      onClick={() => props.onChange(!props.value)}
      id={`wgt-checkbox-${props.id}`}
    />
    <div className="grid gap-1.5 leading-none">
      <label
        htmlFor={`wgt-checkbox-${props.id}`}
        className="text-sm font-medium leading-none text-gray-10 peer-disabled:cursor-not-allowed peer-disabled:opacity-70 dark:text-gray-dark-10"
      >
        {props.label}
      </label>
    </div>
  </div>
);

const widgets: RegistryWidgetsType = {
  CheckboxWidget: CustomCheckbox,
  SelectWidget: CustomSelectWidget,
};

type JSONSchemaFormProps = Omit<
  // copy the props from the original form, and remove the ones we want to implement ourselves
  ComponentProps<typeof Form>,
  "schema" | "templates" | "validator" | "widgets"
> & {
  schema: RJSFSchema;
};

export const JSONSchemaForm = React.forwardRef<Form, JSONSchemaFormProps>(
  ({ schema, className, ...props }, ref) => (
    <Form
      ref={ref}
      schema={schema}
      templates={{
        BaseInputTemplate,
        TitleFieldTemplate,
        ArrayFieldTemplate,
        DescriptionFieldTemplate,
        ErrorListTemplate,
        ButtonTemplates: {
          SubmitButton,
        },
      }}
      validator={validator}
      widgets={widgets}
      {...props}
      className={twMergeClsx("p-1", className)}
    />
  )
);

JSONSchemaForm.displayName = "JSONSchemaForm";
