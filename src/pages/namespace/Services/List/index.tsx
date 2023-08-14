import { Boxes } from "lucide-react";
import { Card } from "~/design/Card";

const ServicesListPage = () => (
  <div className="flex grow flex-col gap-y-4 p-5">
    <h3 className="flex items-center gap-x-2 font-bold">
      <Boxes className="h-5" />
      Services List
    </h3>
    <Card></Card>
  </div>
);

export default ServicesListPage;
