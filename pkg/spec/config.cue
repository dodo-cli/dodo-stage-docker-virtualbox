package spec

include: [...#Include] | *[]

#Include: {
  file: string
}

stages: [string]: #Stage

#Stage: {
  name?:     string
  type:      string
  options:   #Options
  ...
}

#Options: {
  modify:    [...string] | *[]
  provision: [...string] | *[]
}
