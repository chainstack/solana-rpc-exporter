package main

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"sort"
	"testing"
)

type (
	LV struct {
		labels []string
		value  float64
	}
	collectionTest struct {
		Name             string
		ExpectedResponse string
	}
)

func NewLV(value float64, labels ...string) LV {
	return LV{labels, value}
}

func TestNewGaugeDesc(t *testing.T) {
	desc := NewGaugeDesc(
		"test_metric",
		"Test metric description",
		"label1", "label2",
	)

	assert.Equal(t, "test_metric", desc.Name)
	assert.Equal(t, "Test metric description", desc.Help)
	assert.Equal(t, []string{"label1", "label2"}, desc.VariableLabels)
	assert.NotNil(t, desc.Desc)
}

func TestGaugeDesc_MustNewConstMetric(t *testing.T) {
	desc := NewGaugeDesc(
		"test_metric",
		"Test description",
		"label1",
	)

	metric := desc.MustNewConstMetric(42.0, "value1")
	assert.NotNil(t, metric)
	assert.NotNil(t, metric.Desc())
}

func TestGaugeDesc_InvalidLabelCount(t *testing.T) {
	desc := NewGaugeDesc(
		"test_metric",
		"Test description",
		"label1", "label2",
	)

	assert.Panics(t, func() {
		desc.MustNewConstMetric(42.0, "value1") // Should panic - missing label2
	})
}

func TestGaugeDesc_NewInvalidMetric(t *testing.T) {
	desc := NewGaugeDesc("test_metric", "Test description")
	testError := fmt.Errorf("test error")

	metric := desc.NewInvalidMetric(testError)
	assert.NotNil(t, metric)
}

func (c *GaugeDesc) expectedCollection(labeledValues ...LV) string {
	helpLine := fmt.Sprintf("# HELP %s %s", c.Name, c.Help)
	typeLine := fmt.Sprintf("# TYPE %s gauge", c.Name)
	result := fmt.Sprintf("%s\n%s", helpLine, typeLine)

	sortedVariableLabels := make([]string, len(c.VariableLabels))
	copy(sortedVariableLabels, c.VariableLabels)
	sort.Strings(sortedVariableLabels)

	for _, lv := range labeledValues {
		description := ""
		if len(lv.labels) > 0 {
			for i, label := range lv.labels {
				description += fmt.Sprintf("%s=\"%s\",", sortedVariableLabels[i], label)
			}
			description = fmt.Sprintf("{%s}", description[:len(description)-1])
		}
		result += fmt.Sprintf("\n%s%s %v", c.Name, description, lv.value)
	}
	return "\n" + result + "\n"
}

func (c *GaugeDesc) makeCollectionTest(labeledValues ...LV) collectionTest {
	return collectionTest{Name: c.Name, ExpectedResponse: c.expectedCollection(labeledValues...)}
}
