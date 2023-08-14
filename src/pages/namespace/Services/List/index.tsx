import { Boxes } from "lucide-react";
import Button from "~/design/Button";
import { Card } from "~/design/Card";
import { useDeleteService } from "~/api/services/mutate/delete";
import { useServices } from "~/api/services/query/get";

const ServicesListPage = () => {
  const { data } = useServices();
  const { mutate: deleteService } = useDeleteService();

  return (
    <div className="flex grow flex-col gap-y-4 p-5">
      <h3 className="flex items-center gap-x-2 font-bold">
        <Boxes className="h-5" />
        Services List
      </h3>
      <Card>
        {data?.functions.map((service) => (
          <h1 className="p-2" key={service.info.name}>
            {service.info.name}{" "}
            <Button
              variant="destructive"
              size="sm"
              onClick={() => {
                deleteService({
                  service: service.info.name,
                });
              }}
            >
              Delete
            </Button>
          </h1>
        ))}
      </Card>
    </div>
  );
};

export default ServicesListPage;
