import 'dart:convert';

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

var instanceList = '''{
   "workflowInstances":[
      {
         "id":"demo-jcmxg/secrets-example/UzUSOV",
         "status":"pending",
         "beginTime":{
            "seconds":"1615850668",
            "nanos":200317000
         }
      },
      {
         "id":"demo-jcmxg/helloworld/xhjNnh",
         "status":"complete",
         "beginTime":{
            "seconds":"1615850657",
            "nanos":31989000
         }
      },
      {
         "id":"demo-jcmxg/helloworld/PgliqQ",
         "status":"complete",
         "beginTime":{
            "seconds":"1615850651",
            "nanos":806744000
         }
      }
   ],
   "offset":0,
   "limit":10
}
''';

List<Instance> jsonStrToInstanceList(String jsonString) {
  Map<String, dynamic> jsonData = jsonDecode(jsonString);
  final List<Map<String, dynamic>> instances =
      jsonData['workflowInstances'].cast<Map<String, dynamic>>();
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
        'id': '${namespace}/${workflow}/${instanceID}',
        'status': status,
      };
}

var namespaceList = '''{
	"limit": 10,
	"offset": 0,
	"total": 2,
	"data": [
		"demo-jcmxg",
		"komkmk"
	]
}
''';

List<String> jsonStrToNamespaceList(String jsonString) {
  Map<String, dynamic> nsData = jsonDecode(jsonString);
  final List<String> namespaces = nsData['data'].cast<String>();
  return namespaces;
}
