package stacks

// TODO how to represent the different parameters that are system, stack paramters from parameters/{env}.yaml and user supplied required and optional parameters?
// does it matter who supplies them? as long as they are supplied when we call helmfile?
// maybe it does to figure out where we are not supplying them. its not a users fault if we don't specify a system parameter

// TODO is it ok to treat values as strings? using types might be nice

// TODO kubernetes image pull policy should be a valid one. can I use kubernetes to constrain that?

#parameter: {
  value: string
}

#stack: {
    stackName: string
    parameters: [string]: #parameter
}

#dhis2: #stack & {
    stackName: "dhis2"
    parameters: {
       "DATABASE_ID": {}
       "IMAGE_REPOSITORY"?: {
            value: string | *"core"
       }
       "IMAGE_TAG"?: {
            value: string | *"2.39.0"
       }
       "IMAGE_PULL_POLICY"?: {
            value: *"IfNotPresent" | "Always" | "Never"
       }
       "GOOGLE_AUTH_PROJECT_ID"?: {}
    }
}


// These are examples of values that will be merged with the above definition

// instance of dhis2
#dhis2 & {
    parameters: {
     "DATABASE_ID": {
        value: "1"
     }
    }
}

// invalid instance of dhis2
// #dhis2 & {
//     requiredParameters: [
//         {
//             name: "DATABASE_ID"
//             value: 1
//         },
//     ]
// }
