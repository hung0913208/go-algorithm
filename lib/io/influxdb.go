package io

import (
    "context"

    db "github.com/influxdata/influxdb-client-go/v2"
    api "github.com/influxdata/influxdb-client-go/v2/api"
    write "github.com/influxdata/influxdb-client-go/v2/api/write"
)

type influxdbImpl struct {
    client db.Client
    api    api.WriteAPIBlocking
}

var _ Writer = &influxdbImpl{}

func NewInfluxdb(uri, token, organization, bucket string) Writer {
    client := db.NewClient(uri, token)

    return &influxdbImpl{
        client: client,
        api:    client.WriteAPIBlocking(organization, bucket),
    }
}

func (self *influxdbImpl) Write(points []interface{}) error {
    serial := make([]*write.Point, 0)
    for _, point := range points {
        influxPoint, ok := point.(*write.Point)
        if !ok {
            continue
        }

        serial = append(serial, influxPoint)
    }

    return self.api.WritePoint(context.Background(), serial...)
}

func (self *influxdbImpl) Flush() {
}
