import {
  DialogClose,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "~/design/Dialog";
import { Diamond, PlusCircle } from "lucide-react";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "~/design/Select";
import {
  ServiceFormSchema,
  ServiceFormSchemaType,
} from "~/api/services/schema/services";
import { SizeSchema, SizeSchemaType } from "~/api/services/schema";
import { SubmitHandler, useForm } from "react-hook-form";

import Button from "~/design/Button";
import FormErrors from "~/componentsNext/FormErrors";
import Input from "~/design/Input";
import { Slider } from "~/design/Slider";
import { useCreateService } from "~/api/services/mutate/createService";
import { useServices } from "~/api/services/query/getAll";
import { useTranslation } from "react-i18next";
import { zodResolver } from "@hookform/resolvers/zod";

const availableSizes: SizeSchemaType[] = [0, 1, 2];

const CreateService = ({
  close,
  unallowedNames,
}: {
  close: () => void;
  unallowedNames?: string[];
}) => {
  const { t } = useTranslation();

  const { data } = useServices({});
  const { mutate: createService, isLoading } = useCreateService({
    onSuccess: () => {
      close();
    },
  });

  const {
    register,
    handleSubmit,
    watch,
    setValue,
    formState: { isDirty, errors, isValid, isSubmitted },
  } = useForm<ServiceFormSchemaType>({
    defaultValues: {
      minscale: 0,
      size: 1,
    },
    resolver: zodResolver(
      ServiceFormSchema.refine(
        (x) => !(unallowedNames ?? []).some((n) => n === x.name),
        {
          path: ["name"],
          message: t("pages.services.create.nameAlreadyExists"),
        }
      )
    ),
  });

  const onSubmit: SubmitHandler<ServiceFormSchemaType> = ({
    name,
    cmd,
    image,
    minscale,
    size,
  }) => {
    createService({
      name,
      cmd,
      image,
      minscale,
      size,
    });
  };

  // you can not submit if the form has not changed or if there are any errors and
  // you have already submitted the form (errors will first show up after submit)
  const disableSubmit = !isDirty || (isSubmitted && !isValid);

  const formId = `new-service`;

  const maxScale = data?.config.maxscale;
  if (maxScale === undefined) return null;

  const size = watch("size");
  const minscale = watch("minscale");

  return (
    <>
      <DialogHeader>
        <DialogTitle>
          <Diamond /> {t("pages.services.create.title")}
        </DialogTitle>
      </DialogHeader>

      <div className="my-3">
        <FormErrors errors={errors} className="mb-5" />
        <form
          id={formId}
          onSubmit={handleSubmit(onSubmit)}
          className="flex flex-col space-y-5"
        >
          <fieldset className="flex items-center gap-5">
            <label className="w-[90px] text-right text-[14px]" htmlFor="name">
              {t("pages.services.create.nameLabel")}
            </label>
            <Input
              id="name"
              placeholder={t("pages.services.create.namePlaceholder")}
              {...register("name")}
              autoComplete="off"
            />
          </fieldset>
          <fieldset className="flex items-center gap-5">
            <label className="w-[90px] text-right text-[14px]" htmlFor="image">
              {t("pages.services.create.imageLabel")}
            </label>
            <Input
              id="image"
              placeholder={t("pages.services.create.imagePlaceholder")}
              {...register("image")}
            />
          </fieldset>
          <fieldset className="flex items-center gap-5">
            <label className="w-[90px] text-right text-[14px]" htmlFor="scale">
              {t("pages.services.create.scaleLabel")}
            </label>
            <div className="flex w-full gap-5">
              <Input className="w-12" readOnly value={minscale} disabled />
              <Slider
                id="scale"
                step={1}
                min={0}
                max={maxScale}
                value={[minscale ?? 0]}
                onValueChange={(e) => {
                  const newValue = e[0];
                  newValue !== undefined && setValue("minscale", newValue);
                }}
              />
            </div>
          </fieldset>
          <fieldset className="flex items-center gap-5">
            <label className="w-[90px] text-right text-[14px]" htmlFor="size">
              {t("pages.services.create.sizeLabel")}
            </label>
            <Select
              value={`${size}`}
              onValueChange={(value) => {
                const sizeParsed = SizeSchema.safeParse(parseInt(value));
                if (sizeParsed.success) {
                  setValue("size", sizeParsed.data);
                }
              }}
            >
              <SelectTrigger variant="outline" className="w-full" id="size">
                <SelectValue
                  placeholder={t("pages.services.create.sizePlaceholder")}
                />
              </SelectTrigger>
              <SelectContent>
                {availableSizes.map((size) => (
                  <SelectItem key={size} value={`${size}`}>
                    {t(`pages.services.create.sizeValues.${size}`)}
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
          </fieldset>
          <fieldset className="flex items-center gap-5">
            <label className="w-[90px] text-right text-[14px]" htmlFor="cmd">
              {t("pages.services.create.cmdLabel")}
            </label>
            <Input
              id="cmd"
              placeholder={t("pages.services.create.cmdPlaceholder")}
              {...register("cmd")}
            />
          </fieldset>
        </form>
      </div>
      <DialogFooter>
        <DialogClose asChild>
          <Button variant="ghost">
            {t("pages.services.create.cancelBtn")}
          </Button>
        </DialogClose>
        <Button
          type="submit"
          disabled={disableSubmit}
          loading={isLoading}
          form={formId}
        >
          {!isLoading && <PlusCircle />}
          {t("pages.services.create.createBtn")}
        </Button>
      </DialogFooter>
    </>
  );
};

export default CreateService;
