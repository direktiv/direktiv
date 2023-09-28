import {
  DialogClose,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "~/design/Dialog";
import { Diamond, PlusCircle } from "lucide-react";
import {
  RevisionFormSchema,
  RevisionFormSchemaType,
} from "~/api/services/schema/revisions";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "~/design/Select";
import { SubmitHandler, useForm } from "react-hook-form";

import Button from "~/design/Button";
import FormErrors from "~/componentsNext/FormErrors";
import Input from "~/design/Input";
import { SizeSchema } from "~/api/services/schema";
import { Slider } from "~/design/Slider";
import { useCreateServiceRevision } from "~/api/services/mutate/createRevision";
import { useServiceDetails } from "~/api/services/query/getDetails";
import { useTranslation } from "react-i18next";
import { zodResolver } from "@hookform/resolvers/zod";

const availableSizes = [0, 1, 2] as const;

const CreateRevision = ({
  service,
  defaultValues,
  close,
}: {
  service: string;
  close: () => void;
  defaultValues?: RevisionFormSchemaType;
}) => {
  const { t } = useTranslation();

  const { data } = useServiceDetails({ service });
  const { mutate: createServiceRevision, isLoading } = useCreateServiceRevision(
    {
      onSuccess: () => {
        close();
      },
    }
  );

  const {
    register,
    handleSubmit,
    watch,
    setValue,
    formState: { errors },
  } = useForm<RevisionFormSchemaType>({
    defaultValues: defaultValues ?? {
      minscale: 0,
      size: 1,
    },
    resolver: zodResolver(
      RevisionFormSchema.refine(
        (x) => {
          if (defaultValues) {
            return Object.keys(defaultValues).some((key) => {
              const typedKey = key as keyof typeof defaultValues;
              return x[typedKey] !== defaultValues[typedKey];
            });
          }
          return true;
        },
        {
          message: t("pages.services.revision.create.noChanges"),
        }
        /**
         * when no default values are available, it problably means that
         * there is no previous revision. In this case, it doesn't make
         * sense to let the user create one
         */
      ).refine(() => !!defaultValues, {
        message: t("pages.services.revision.create.noDefaultValues"),
      })
    ),
  });

  const onSubmit: SubmitHandler<RevisionFormSchemaType> = ({
    cmd,
    image,
    minscale,
    size,
  }) => {
    createServiceRevision({
      service,
      payload: {
        cmd,
        image,
        minscale,
        size,
      },
    });
  };

  const disableSubmit = isLoading;

  const formId = `new-service-revision`;

  const maxScale = data?.config?.maxscale;
  if (maxScale === undefined) return null;

  const size = watch("size");
  const minscale = watch("minscale");

  return (
    <>
      <DialogHeader>
        <DialogTitle>
          <Diamond /> {t("pages.services.revision.create.title")}
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
            <label className="w-[90px] text-right text-[14px]" htmlFor="image">
              {t("pages.services.revision.create.imageLabel")}
            </label>
            <Input
              id="image"
              placeholder={t("pages.services.revision.create.imagePlaceholder")}
              {...register("image")}
            />
          </fieldset>
          <fieldset className="flex items-center gap-5">
            <label className="w-[90px] text-right text-[14px]" htmlFor="scale">
              {t("pages.services.revision.create.scaleLabel")}
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
              {t("pages.services.revision.create.sizeLabel")}
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
                  placeholder={t(
                    "pages.services.revision.create.sizePlaceholder"
                  )}
                />
              </SelectTrigger>
              <SelectContent>
                {availableSizes.map((size) => (
                  <SelectItem key={size} value={`${size}`}>
                    {t(`pages.services.revision.create.sizeValues.${size}`)}
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
          </fieldset>
          <fieldset className="flex items-center gap-5">
            <label className="w-[90px] text-right text-[14px]" htmlFor="cmd">
              {t("pages.services.revision.create.cmdLabel")}
            </label>
            <Input
              id="cmd"
              placeholder={t("pages.services.revision.create.cmdPlaceholder")}
              {...register("cmd")}
            />
          </fieldset>
        </form>
      </div>
      <DialogFooter>
        <DialogClose asChild>
          <Button variant="ghost">
            {t("pages.services.revision.create.cancelBtn")}
          </Button>
        </DialogClose>
        <Button
          type="submit"
          disabled={disableSubmit}
          loading={isLoading}
          form={formId}
        >
          {!isLoading && <PlusCircle />}
          {t("pages.services.revision.create.createBtn")}
        </Button>
      </DialogFooter>
    </>
  );
};

export default CreateRevision;
