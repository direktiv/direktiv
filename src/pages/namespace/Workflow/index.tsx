import { GitCommit, GitMerge, PieChart, Play, Settings } from "lucide-react";

import Button from "../../../componentsNext/Button";
import { FC } from "react";
import { RxChevronDown } from "react-icons/rx";
import clsx from "clsx";

const tabs = [
  { name: "Overview", href: "#", icon: PieChart, current: true },
  { name: "Active Revisions", href: "#", icon: GitCommit, current: false },
  { name: "Revisions", href: "#", icon: GitMerge, current: false },
  { name: "Settings", href: "#", icon: Settings, current: false },
];
const WorkflowPage: FC = () => (
  <div className="space-y-5 border-b border-gray-gray5 bg-base-200 p-5 pb-0 dark:border-grayDark-gray5">
    <div className="flex flex-col max-sm:space-y-4 sm:flex-row sm:items-center sm:justify-between ">
      <h3 className="flex items-center gap-x-2 font-bold text-primary-500">
        <Play className="h-5" />
        workflow.yml
      </h3>
      <Button color="primary" size="sm">
        Actions <RxChevronDown />
      </Button>
    </div>
    <div>
      <nav className="-mb-px flex space-x-8">
        {tabs.map((tab) => (
          <a
            key={tab.name}
            href={tab.href}
            className={clsx(
              tab.current
                ? "border-primary-500 text-primary-500"
                : "border-transparent text-gray-gray11 hover:border-gray-gray8 hover:text-gray-gray12 dark:hover:border-grayDark-gray8 dark:hover:text-grayDark-gray12",
              "flex items-center gap-x-2 whitespace-nowrap border-b-2 px-1 pb-4 text-sm font-medium"
            )}
            aria-current={tab.current ? "page" : undefined}
          >
            <tab.icon aria-hidden="true" className="h-4 w-auto" /> {tab.name}
          </a>
        ))}
      </nav>
    </div>
  </div>
);

export default WorkflowPage;
