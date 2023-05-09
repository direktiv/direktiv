import {
  DialogClose,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "../../../../../design/Dialog";
import { Play, PlusCircle } from "lucide-react";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "../../../../../design/Select";
import { SubmitHandler, useForm } from "react-hook-form";

import Alert from "../../../../../design/Alert";
import Button from "../../../../../design/Button";
import Input from "../../../../../design/Input";
import { Textarea } from "../../../../../design/TextArea";
import { fileNameSchema } from "../../../../../api/tree/schema";
import { pages } from "../../../../../util/router/pages";
import { useCreateWorkflow } from "../../../../../api/tree/mutate/createWorkflow";
import { useNamespace } from "../../../../../util/store/namespace";
import { useNavigate } from "react-router-dom";
import workflowTemplates from "./templates";
import { z } from "zod";
import { zodResolver } from "@hookform/resolvers/zod";

type FormInput = {
  name: string;
  fileContent: string;
};

const defaultWorkflowTemplate = workflowTemplates[0];

const NewWorkflow = ({
  path,
  close,
  unallowedNames,
}: {
  path?: string;
  close: () => void;
  unallowedNames: string[];
}) => {
  const namespace = useNamespace();
  const navigate = useNavigate();
  const {
    register,
    handleSubmit,
    setValue,
    formState: { isDirty, errors, isValid, isSubmitted },
  } = useForm<FormInput>({
    resolver: zodResolver(
      z.object({
        name: fileNameSchema.and(
          z.string().refine((name) => !unallowedNames.some((n) => n === name), {
            message: "The name already exists",
          })
        ),
        fileContent: z.string(),
      })
    ),
    defaultValues: {
      fileContent: defaultWorkflowTemplate.data,
    },
  });

  const { mutate: createWorkflow, isLoading } = useCreateWorkflow({
    onSuccess: (data) => {
      namespace &&
        navigate(
          pages.explorer.createHref({
            namespace,
            path: data.node.path,
            subpage: "workflow",
          })
        );
      close();
    },
  });

  const onSubmit: SubmitHandler<FormInput> = ({ name, fileContent }) => {
    createWorkflow({ path, name, fileContent });
  };

  // you can not submit if the form has not changed or if there are any errors and
  // you have already submitted the form (errors will first show up after submit)
  const disableSubmit = !isDirty || (isSubmitted && !isValid);

  const formId = `new-worfklow-${path}`;
  return (
    <>
      <DialogHeader>
        <DialogTitle>
          <Play /> Create a new Workflow
        </DialogTitle>
      </DialogHeader>

      <div className="my-3">
        {!!errors.name && (
          <Alert variant="error" className="mb-5">
            <p>{errors.name.message}</p>
          </Alert>
        )}
        <form
          id={formId}
          onSubmit={handleSubmit(onSubmit)}
          className="flex flex-col space-y-5"
        >
          <fieldset className="flex items-center gap-5">
            <label className="w-[150px] text-right text-[15px]" htmlFor="name">
              Name
            </label>
            <Input
              data-testid="new-workflow-name"
              id="name"
              placeholder="workflow-name"
              {...register("name")}
            />
          </fieldset>
          <fieldset className="flex items-center gap-5">
            <label
              className="w-[150px] text-right text-[15px]"
              htmlFor="template"
            >
              template
            </label>
            <Select
              onValueChange={(value) => {
                const matchingWf = workflowTemplates.find(
                  (t) => t.name === value
                );
                if (matchingWf) setValue("fileContent", matchingWf.data);
              }}
            >
              <SelectTrigger id="template" variant="outline">
                <SelectValue
                  placeholder={defaultWorkflowTemplate.name}
                  defaultValue={defaultWorkflowTemplate.data}
                />
              </SelectTrigger>
              <SelectContent>
                {workflowTemplates.map((t) => (
                  <SelectItem value={t.name} key={t.name}>
                    {t.name}
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
          </fieldset>
          <fieldset className="flex items-start gap-5">
            <Textarea
              className="h-96"
              data-testid="new-workflow-editor"
              {...register("fileContent")}
            />
          </fieldset>
        </form>
      </div>
      <DialogFooter>
        <DialogClose asChild>
          <Button variant="ghost">Cancel</Button>
        </DialogClose>
        <Button
          data-testid="new-workflow-submit"
          type="submit"
          disabled={disableSubmit}
          loading={isLoading}
          form={formId}
        >
          {!isLoading && <PlusCircle />}
          Create
        </Button>
      </DialogFooter>
    </>
  );
};

export default NewWorkflow;
