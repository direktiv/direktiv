import { z } from "zod";

/**
 * example response
 * 
 * {
    "results": [
        {
            "metric": {
                "__name__": "direktiv_workflows_success_total",
                "app_kubernetes_io_instance": "direktiv",
                "app_kubernetes_io_name": "direktiv",
                "direktiv_namespace": "sebxian",
                "direktiv_tenant": "sebxian",
                "direktiv_workflow": "/greeting/generator.yaml",
                "instance": "10.42.0.186:2112",
                "job": "kubernetes-pods",
                "namespace": "default",
                "pod": "direktiv-flow-7c6c58db6f-zpm9c",
                "pod_template_hash": "7c6c58db6f"
            },
            "value": [
                1693569142.821,
                "5"
            ]
        }
    ],
    "warnings": null
  }
 */

export const MetricsListSchema = z.object({
  results: z.array(
    z.object({
      value: z.tuple([z.number(), z.string()]),
    })
  ),
});

export type MetricsListSchemaType = z.infer<typeof MetricsListSchema>;
