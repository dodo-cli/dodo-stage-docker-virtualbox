package spec

stages: [string]: #Stage

#Stage: {
  name?:     string
  type:      string
  options:   #Options
  ...
}

#Options: {
  modify:        [...string] | *[]
  provision:     [...string] | *[]
  stagehandUrl?: string
}
