import { CalendarDays, Info } from "lucide-react";
import { Popover, PopoverContent, PopoverTrigger } from "~/design/Popover";

import Button from "~/design/Button";
import { Fragment } from "react";
import { useTranslation } from "react-i18next";

const useExamples = () => {
  const { t } = useTranslation();

  return [
    {
      title: t("pages.permissions.durationHint.P1Y"),
      duration: "P1Y",
    },
    {
      title: t("pages.permissions.durationHint.PT48H"),
      duration: "PT48H",
    },
    {
      title: "1 year, 2 months, and 15 days",
      duration: "P1Y2M15D",
    },
  ] as const;
};

const DurationHint = ({
  onDurationSelect,
}: {
  onDurationSelect: (duration: string) => void;
}) => {
  const { t } = useTranslation();

  const examples = useExamples();

  return (
    <Popover>
      <PopoverTrigger asChild>
        <Button icon variant="ghost" type="button">
          <Info />
        </Button>
      </PopoverTrigger>
      <PopoverContent asChild>
        <div className="flex w-96 flex-col gap-3 p-5">
          <h3 className="flex items-center gap-x-2 font-bold">
            <CalendarDays className="h-5" />
            {t("pages.permissions.durationHint.title")}
          </h3>
          <p>{t("pages.permissions.durationHint.description")}</p>
          <b>{t("pages.permissions.durationHint.examples")}</b>
          <div className="grid grid-cols-[auto_1fr] gap-x-5 gap-y-3">
            {examples.map(({ title, duration }) => (
              <Fragment key={duration}>
                <div>{title}</div>
                <Button
                  variant="outline"
                  size="sm"
                  onClick={() => onDurationSelect(duration)}
                >
                  {duration}
                </Button>
              </Fragment>
            ))}
          </div>
        </div>
      </PopoverContent>
    </Popover>
  );
};

export default DurationHint;
