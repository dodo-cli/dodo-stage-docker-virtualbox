package spec

stages: [string]: #Stage

#Stage: {
  name?:     string
  type:      string

  if type == "virtualbox" {
    box:       #VBBox
    resources: #VBResources
    options:   [...string] | *[]
  }

  ...
}

#VBBox: {
  user:          string
  name:          string
  version:       string
  access_token?: string
}

#VBResources: {
  cpu:     int
  memory:  string
  volumes: #VBVolumes | [...#VBVolume] | *[]
  usb:     #VBUSBFilters |[...#VBUSBFilter] | *[]
}

#VBVolumes: [string]: #VBVolume

#VBVolume: {
  size: string
}

#VBUSBFilters: [string]: #VBUSBFilter

#VBUSBFilter: {
  name:      string
  vendorid:  string
  productid: string
}
