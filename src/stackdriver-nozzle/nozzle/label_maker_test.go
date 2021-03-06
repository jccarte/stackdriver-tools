/*
 * Copyright 2017 Google Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *    http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package nozzle

import (
	"time"

	"github.com/cloudfoundry-community/stackdriver-tools/src/stackdriver-nozzle/cloudfoundry"
	"github.com/cloudfoundry-community/stackdriver-tools/src/stackdriver-nozzle/mocks"
	"github.com/cloudfoundry/sonde-go/events"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

const foundation = "bosh-foundation"

var _ = Describe("LabelMaker", func() {
	var (
		subject  LabelMaker
		envelope *events.Envelope
	)

	BeforeEach(func() {
		subject = NewLabelMaker(cloudfoundry.NullAppInfoRepository(), foundation)
	})

	It("makes labels from envelopes", func() {
		origin := "cool-origin"
		eventType := events.Envelope_HttpStartStop
		timestamp := time.Now().UnixNano()
		deployment := "neat-deployment"
		job := "some-job"
		index := "an-index"
		ip := "192.168.1.1"
		tags := map[string]string{
			"foo": "bar",
			"bar": "foo",
		}

		envelope = &events.Envelope{
			Origin:     &origin,
			EventType:  &eventType,
			Timestamp:  &timestamp,
			Deployment: &deployment,
			Job:        &job,
			Index:      &index,
			Ip:         &ip,
			Tags:       tags,
		}

		metricLabels := subject.MetricLabels(envelope, false)
		logLabels := subject.LogLabels(envelope)

		Expect(metricLabels).To(Equal(map[string]string{
			"foundation": foundation,
			"job":        job,
			"index":      index,
			"tags":       "bar=foo,foo=bar",
		}))
		Expect(logLabels).To(Equal(map[string]string{
			"foundation": foundation,
			"job":        job,
			"index":      index,
			"tags":       "bar=foo,foo=bar",
			"origin":     origin,
			"eventType":  "HttpStartStop",
		}))
	})

	It("ignores empty fields", func() {
		origin := "cool-origin"
		eventType := events.Envelope_HttpStartStop
		timestamp := time.Now().UnixNano()
		job := "some-job"
		index := "an-index"
		tags := map[string]string{
			"foo": "bar",
		}

		envelope := &events.Envelope{
			Origin:     &origin,
			EventType:  &eventType,
			Timestamp:  &timestamp,
			Deployment: nil,
			Job:        &job,
			Index:      &index,
			Ip:         nil,
			Tags:       tags,
		}

		labels := subject.MetricLabels(envelope, false)

		Expect(labels).To(Equal(map[string]string{
			"foundation": foundation,
			"job":        job,
			"index":      index,
			"tags":       "foo=bar",
		}))
	})

	Context("Metadata", func() {
		var (
			appGUID = "f47ac10b-58cc-4372-a567-0e02b2c3d479"
			low     = uint64(0x7243cc580bc17af4)
			high    = uint64(0x79d4c3b2020e67a5)
			appID   = events.UUID{Low: &low, High: &high}
		)

		Context("application metadata", func() {
			var (
				appInfoRepository *mocks.AppInfoRepository
			)

			BeforeEach(func() {
				appInfoRepository = &mocks.AppInfoRepository{
					AppInfoMap: map[string]cloudfoundry.AppInfo{},
				}
				subject = NewLabelMaker(appInfoRepository, foundation)
			})

			Context("for a LogMessage", func() {
				var (
					eventType    = events.Envelope_LogMessage
					event        *events.LogMessage
					envelope     *events.Envelope
					spaceGUID    = "2ab560c3-3f21-45e0-9452-d748ff3a15e9"
					orgGUID      = "b494fb47-3c44-4a98-9a08-d839ec5c799b"
					instanceGUID = "301f96f1-97f8-42f8-aa98-6f13ea1f0b87"
				)

				BeforeEach(func() {
					event = &events.LogMessage{
						AppId:          &appGUID,
						SourceInstance: &instanceGUID,
					}
					envelope = &events.Envelope{
						EventType:  &eventType,
						LogMessage: event,
					}
				})

				It("adds fields for a resolved app", func() {
					app := cloudfoundry.AppInfo{
						AppName:   "MyApp",
						SpaceName: "MySpace",
						SpaceGUID: spaceGUID,
						OrgName:   "MyOrg",
						OrgGUID:   orgGUID,
					}

					appInfoRepository.AppInfoMap[appGUID] = app

					labels := subject.MetricLabels(envelope, false)

					Expect(labels).To(HaveKeyWithValue("applicationPath",
						"/MyOrg/MySpace/MyApp"))
					Expect(labels).To(HaveKeyWithValue("instanceIndex",
						instanceGUID))
				})

				It("doesn't add fields for an unresolved app", func() {
					labels := subject.MetricLabels(envelope, false)

					Expect(labels).NotTo(HaveKey("applicationPath"))
				})
			})
			Context("for an HttpStartStop", func() {
				var (
					eventType    = events.Envelope_HttpStartStop
					event        *events.HttpStartStop
					envelope     *events.Envelope
					spaceGUID    = "2ab560c3-3f21-45e0-9452-d748ff3a15e9"
					orgGUID      = "b494fb47-3c44-4a98-9a08-d839ec5c799b"
					instanceIdx  = int32(1)
					instanceGUID = "485a10c1-917f-4d89-a98f-dc539ba14dfd"
				)

				BeforeEach(func() {
					event = &events.HttpStartStop{
						ApplicationId: &appID,
						InstanceIndex: &instanceIdx,
						InstanceId:    &instanceGUID,
					}
					envelope = &events.Envelope{
						EventType:     &eventType,
						HttpStartStop: event,
					}
				})

				It("adds fields for a resolved app", func() {
					app := cloudfoundry.AppInfo{
						AppName:   "MyApp",
						SpaceName: "MySpace",
						SpaceGUID: spaceGUID,
						OrgName:   "MyOrg",
						OrgGUID:   orgGUID,
					}

					appInfoRepository.AppInfoMap[appGUID] = app

					labels := subject.MetricLabels(envelope, false)

					Expect(labels).To(HaveKeyWithValue("applicationPath",
						"/MyOrg/MySpace/MyApp"))
					Expect(labels).To(HaveKeyWithValue("instanceIndex", "1"))
				})

				It("falls back to instance UUID", func() {
					app := cloudfoundry.AppInfo{
						AppName:   "MyApp",
						SpaceName: "MySpace",
						SpaceGUID: spaceGUID,
						OrgName:   "MyOrg",
						OrgGUID:   orgGUID,
					}

					appInfoRepository.AppInfoMap[appGUID] = app

					envelope.HttpStartStop.InstanceIndex = nil
					labels := subject.MetricLabels(envelope, false)

					Expect(labels).To(HaveKeyWithValue("applicationPath",
						"/MyOrg/MySpace/MyApp"))
					Expect(labels).To(HaveKeyWithValue("instanceIndex", instanceGUID))
				})

				It("doesn't add fields for an unresolved app", func() {
					labels := subject.MetricLabels(envelope, false)

					Expect(labels).NotTo(HaveKey("applicationPath"))
				})
			})
		})
	})
})
