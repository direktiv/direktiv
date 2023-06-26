import {
  Table,
  TableBody,
  TableHead,
  TableHeaderCell,
  TableRow,
} from "~/design/Table";

import { Boxes } from "lucide-react";
import { Card } from "~/design/Card";
import Row from "./Row";
import { useInstances } from "~/api/instances/query/get";

const InstancesListPage = () => {
  const { data } = useInstances({ limit: 10, offset: 0 });

  return (
    <div className="flex grow flex-col gap-y-4 p-5">
      <h3 className="flex items-center gap-x-2 font-bold text-gray-10 dark:text-gray-dark-10">
        <Boxes className="h-5" />
        Recently executed instances
      </h3>
      <Card>
        <Table>
          <TableHead>
            <TableRow>
              <TableHeaderCell>name</TableHeaderCell>
              <TableHeaderCell>revision id</TableHeaderCell>
              <TableHeaderCell>state</TableHeaderCell>
              <TableHeaderCell>started at</TableHeaderCell>
              <TableHeaderCell>last updated</TableHeaderCell>
            </TableRow>
          </TableHead>
          <TableBody>
            {data?.instances.results.map((instance) => (
              <Row
                instance={instance}
                key={instance.id}
                namespace={data.namespace}
              />
            ))}
          </TableBody>
        </Table>
      </Card>
    </div>
  );
};

export default InstancesListPage;
