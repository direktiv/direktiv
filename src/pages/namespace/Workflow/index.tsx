import { GitCommit, GitMerge, PieChart, Play, Settings } from "lucide-react";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "../../../design/Tabs";

import Button from "../../../design/Button";
import { FC } from "react";
import { RxChevronDown } from "react-icons/rx";

const WorkflowPage: FC = () => (
  <div className="space-y-5 border-b border-gray-5 bg-gray-2 p-5 pb-0 dark:border-gray-dark-5">
    <div className="flex flex-col max-sm:space-y-4 sm:flex-row sm:items-center sm:justify-between">
      <h3 className="flex items-center gap-x-2 font-bold text-primary-500">
        <Play className="h-5" />
        workflow.yml
      </h3>
      <Button variant="primary">
        Actions <RxChevronDown />
      </Button>
    </div>
    <div>
      <nav className="-mb-px flex space-x-8">
        <Tabs defaultValue="overview">
          <TabsList>
            <TabsTrigger value="overview">
              <PieChart aria-hidden="true" />
              Overview
            </TabsTrigger>
            <TabsTrigger value="active-rev">
              <GitCommit aria-hidden="true" />
              Active Revisions
            </TabsTrigger>
            <TabsTrigger value="revisions">
              <GitMerge aria-hidden="true" />
              Revisions
            </TabsTrigger>
            <TabsTrigger value="settings">
              <Settings aria-hidden="true" />
              Settings
            </TabsTrigger>
          </TabsList>
          <TabsContent value="account">
            <p className="text-sm text-gray-8 dark:text-gray-dark-8">
              Make changes to your account here. Click save when you&apos;re
              done.
            </p>
          </TabsContent>
          <TabsContent value="password">
            <p className="text-sm text-gray-8 dark:text-gray-dark-8">
              Change your password here. After saving, you&apos;ll be logged
              out.
            </p>
          </TabsContent>
          <TabsContent value="third">
            <p className="text-sm text-gray-8 dark:text-gray-dark-8">
              Your third content here
            </p>
          </TabsContent>
          <TabsContent value="fourth">
            <p className="text-sm text-gray-8 dark:text-gray-dark-8">
              The fourth content comes here
            </p>
          </TabsContent>
        </Tabs>
      </nav>
    </div>
  </div>
);

export default WorkflowPage;
