import 'dart:convert';
import 'package:http/http.dart' as http;

var apiSERVER = "http://localhost:8989";
var workflowList = '''{
   "workflows":[
      {
         "uid":"0b94cbbe-8150-4fda-a6f4-688b45b07eca",
         "id":"get-instances-list",
         "revision":16,
         "active":true,
         "createdAt":{
            "seconds":"1615418485",
            "nanos":916110000
         },
         "description":"Displays a gcp project instances"
      },
      {
         "uid":"1cb113ed-6428-48f7-ac9c-3330390623e1",
         "id":"titty-checkerv2",
         "revision":15,
         "active":true,
         "createdAt":{
            "seconds":"1614917442",
            "nanos":962088000
         },
         "description":"Checks a image and gives it a 🍑 NSFW score"
      },
      {
         "uid":"64b18f2b-a7e9-4177-9dc1-5ee243f99ec0",
         "id":"helloworld",
         "revision":8,
         "active":true,
         "createdAt":{
            "seconds":"1615499112",
            "nanos":781936000
         },
         "description":""
      },
      {
         "uid":"8573d6de-1a91-4f2a-9762-4efe668e2071",
         "id":"scott-dad-jokes",
         "revision":7,
         "active":true,
         "createdAt":{
            "seconds":"1614828109",
            "nanos":747572000
         },
         "description":"🏴󠁧󠁢󠁳󠁣󠁴󠁿 🏴󠁧󠁢󠁳󠁣󠁴󠁿 🏴󠁧󠁢󠁳󠁣󠁴󠁿"
      },
      {
         "uid":"e654dacf-aff8-45ea-9dd3-41cac824700b",
         "id":"secrets-example",
         "revision":19,
         "active":true,
         "createdAt":{
            "seconds":"1614741922",
            "nanos":519579000
         },
         "description":"Example of using secrets in action"
      }
   ],
   "offset":0,
   "limit":10,
   "total":5
}
''';

class Workflow {
  final String uid;
  final String id;
  final int revision;
  final bool active;
  final String createdAt;
  final String description;

  Workflow(this.uid, this.id, this.revision, this.active, this.createdAt,
      this.description);

  factory Workflow.fromJson(Map<String, dynamic> json) {
    String createdAt = "";
    return new Workflow(json["uid"], json["id"], json["revision"],
        json["active"], createdAt, json["description"]);
  }
}

Future<List<Workflow>> fetchWorkflow(String namespace) async {
  final response =
      await http.get('$apiSERVER/api/namespaces/$namespace/workflows');
  if (response.statusCode == 200) {
    return jsonStrToWorkflowList(response.body);
  } else {
    throw Exception('Failed to load Instance List');
  }
}

List<Workflow> jsonStrToWorkflowList(String jsonString) {
  Map<String, dynamic> jsonData = jsonDecode(jsonString);
  List<Map<String, dynamic>> workflows = [];

  if (jsonData.containsKey('workflows')) {
    workflows = jsonData['workflows'].cast<Map<String, dynamic>>();
  }

  List<Workflow> workflowList = [];
  workflows.forEach((element) {
    workflowList.add(Workflow.fromJson(element));
  });
  return workflowList;
}

Future<List<Instance>> fetchNamespaceInstances(String namespace) async {
  final response = await http.get('$apiSERVER/api/instances/$namespace');
  if (response.statusCode == 200) {
    return jsonStrToInstanceList(response.body);
  } else {
    throw Exception('Failed to load Instance List');
  }
}

List<Instance> jsonStrToInstanceList(String jsonString) {
  Map<String, dynamic> jsonData = jsonDecode(jsonString);
  List<Map<String, dynamic>> instances = [];

  if (jsonData.containsKey('workflowInstances')) {
    instances = jsonData['workflowInstances'].cast<Map<String, dynamic>>();
  }

  List<Instance> instanceList = [];
  instances.forEach((element) {
    instanceList.add(Instance.fromJson(element));
  });
  return instanceList;
}

class Instance {
  final String workflow;
  final String namespace;
  final String instanceID;
  final String status;

  Instance(this.workflow, this.namespace, this.instanceID, this.status);

  factory Instance.fromJson(Map<String, dynamic> json) {
    String wfInstance = json["id"].toString();
    List<String> wfiSplit = wfInstance.split("/");
    return new Instance(wfiSplit[0], wfiSplit[1], wfiSplit[2], json["status"]);
  }

  Map<String, dynamic> toJson() => {
        'id': '$namespace/$workflow/$instanceID',
        'status': status,
      };
}

class Namespace {
  String name;
  List<Instance> instances;
  Namespace(this.name, this.instances);
}

Future<List<Namespace>> fetchNamespaces() async {
  final response = await http.get('$apiSERVER/api/namespaces');
  List<Namespace> namespaceList = [];
  if (response.statusCode == 200) {
    final nsList = jsonStrToNamespaceList(response.body);
    for (var ns in nsList) {
      final nsInstances = await fetchNamespaceInstances(ns);
      namespaceList.add(Namespace(ns, nsInstances));
    }
    return namespaceList;
  } else {
    throw Exception('Failed to load Instance List');
  }
}

List<String> jsonStrToNamespaceList(String jsonString) {
  Map<String, dynamic> jsonData = jsonDecode(jsonString);
  final List<Map<String, dynamic>> ns =
      jsonData['namespaces'].cast<Map<String, dynamic>>();
  List<String> namespaceList = [];
  ns.forEach((element) {
    namespaceList.add(element['name']);
  });
  return namespaceList;
}
