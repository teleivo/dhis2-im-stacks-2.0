package stacks

// TODO how to represent the different parameters that are system, stack paramters from parameters/{env}.yaml and user supplied required and optional parameters?
// does it matter who supplies them? as long as they are supplied when we call helmfile?
// maybe it does to figure out where we are not supplying them. its not a users fault if we don't specify a system parameter

// TODO is it ok to treat values as strings? using types might be nice

#parameter: {
  name: string
  value: string
}

#stack: {
    name: string
    parameters: [...#parameter]
}

#dhis2: #stack & {
    name: "dhis2"
    parameters: [
        {
            name: DATABASE_USERNAME,
        },
        {
            name: GOOGLE_AUTH_PROJECT_ID: string | *"",
        }
    ]
}

// instance of dhis2
valid: #dhis2 & {
    parameters: {
        DATABASE_USERNAME: "ivo"
    }
}

// invalid instance of dhis2
// invalid: #dhis2 & {
//     parameters: {
//         DATABASE_USERNAME: 2
//     }
// }
