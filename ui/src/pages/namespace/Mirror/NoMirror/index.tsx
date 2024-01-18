import { Book, FlaskConical, GitCompare } from "lucide-react";

import { Card } from "~/design/Card";
import { useTranslation } from "react-i18next";

const NoMirror = () => {
  const { t } = useTranslation();

  const docsUrl = "https://docs.direktiv.io/environment/git/";
  const examplesUrl = "https://github.com/direktiv/direktiv-examples";

  return (
    <div className="flex grow flex-col gap-y-4 p-6">
      <h3 className="flex items-center gap-x-2 font-bold">
        <GitCompare className="h-5" />
        {t("pages.mirror.noMirror.title")}
      </h3>
      <Card className="flex max-w-xl flex-col gap-5 p-6">
        <p className="text-sm">{t("pages.mirror.noMirror.description1")}</p>
        <p className="text-sm">{t("pages.mirror.noMirror.description2")}</p>

        <div className="group relative flex items-start space-x-3">
          <div className="shrink-0">
            <span className="inline-flex h-10 w-10 items-center justify-center rounded-lg">
              <Book />
            </span>
          </div>
          <div className="min-w-0 flex-1">
            <div className="text-sm font-medium text-gray-11 dark:text-gray-dark-11">
              <a href={docsUrl} target="_blank" rel="noopener noreferrer">
                <span className="absolute inset-0" aria-hidden="true" />
                {t("pages.mirror.noMirror.docs.title")}
              </a>
            </div>
            <p className="text-sm text-gray-9 dark:text-gray-dark-9">
              {t("pages.mirror.noMirror.docs.description")}
            </p>
          </div>
          <div className="shrink-0 self-center"></div>
        </div>

        <div className="group relative flex items-start space-x-3">
          <div className="shrink-0">
            <span className="inline-flex h-10 w-10 items-center justify-center rounded-lg">
              <FlaskConical />
            </span>
          </div>
          <div className="min-w-0 flex-1">
            <div className="text-sm font-medium text-gray-11 dark:text-gray-dark-11">
              <a href={examplesUrl} target="_blank" rel="noopener noreferrer">
                <span className="absolute inset-0" aria-hidden="true" />
                {t("pages.mirror.noMirror.example.title")}
              </a>
            </div>
            <p className="text-sm text-gray-9 dark:text-gray-dark-9">
              {t("pages.mirror.noMirror.example.description")}
            </p>
          </div>
          <div className="shrink-0 self-center"></div>
        </div>
      </Card>
    </div>
  );
};

export default NoMirror;
