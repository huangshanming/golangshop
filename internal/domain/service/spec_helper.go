package service

import (
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strings"

	"goshop/internal/domain/model"
)

type SpecDimension struct {
	Name   string   `json:"name"`
	Values []string `json:"values"`
}

func parseSpecDimensions(raw string) ([]SpecDimension, error) {
	if strings.TrimSpace(raw) == "" {
		return nil, nil
	}
	var dims []SpecDimension
	if err := json.Unmarshal([]byte(raw), &dims); err != nil {
		return nil, fmt.Errorf("规格定义 JSON 无效")
	}
	for i, d := range dims {
		name := strings.TrimSpace(d.Name)
		if name == "" {
			return nil, fmt.Errorf("规格维度名称不能为空")
		}
		dims[i].Name = name
		vals := make([]string, 0, len(d.Values))
		seen := map[string]bool{}
		for _, v := range d.Values {
			v = strings.TrimSpace(v)
			if v == "" || seen[v] {
				continue
			}
			seen[v] = true
			vals = append(vals, v)
		}
		if len(vals) == 0 {
			return nil, fmt.Errorf("规格「%s」至少需要一个值", name)
		}
		dims[i].Values = vals
	}
	return dims, nil
}

func formatSpecSummary(specValues map[string]string, dims []SpecDimension) string {
	if len(specValues) == 0 {
		return ""
	}
	if len(dims) > 0 {
		parts := make([]string, 0, len(dims))
		for _, d := range dims {
			if v, ok := specValues[d.Name]; ok && v != "" {
				parts = append(parts, v)
			}
		}
		if len(parts) > 0 {
			return strings.Join(parts, " / ")
		}
	}
	keys := make([]string, 0, len(specValues))
	for k := range specValues {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	parts := make([]string, 0, len(keys))
	for _, k := range keys {
		parts = append(parts, specValues[k])
	}
	return strings.Join(parts, " / ")
}

func specValuesKey(specValues map[string]string, dims []SpecDimension) string {
	if len(dims) > 0 {
		parts := make([]string, 0, len(dims))
		for _, d := range dims {
			parts = append(parts, d.Name+"="+specValues[d.Name])
		}
		return strings.Join(parts, "|")
	}
	keys := make([]string, 0, len(specValues))
	for k := range specValues {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	parts := make([]string, 0, len(keys))
	for _, k := range keys {
		parts = append(parts, k+"="+specValues[k])
	}
	return strings.Join(parts, "|")
}

func parseSpecValuesJSON(raw string) (map[string]string, error) {
	if strings.TrimSpace(raw) == "" {
		return map[string]string{}, nil
	}
	var m map[string]string
	if err := json.Unmarshal([]byte(raw), &m); err != nil {
		return nil, err
	}
	return m, nil
}

func marshalSpecValues(m map[string]string) string {
	if len(m) == 0 {
		return ""
	}
	b, _ := json.Marshal(m)
	return string(b)
}

func marshalSpecs(dims []SpecDimension) string {
	if len(dims) == 0 {
		return ""
	}
	b, _ := json.Marshal(dims)
	return string(b)
}

func buildSKUsFromInput(specsRaw string, items []SKUInputItem) ([]model.ProductSKU, error) {
	dims, err := parseSpecDimensions(specsRaw)
	if err != nil {
		return nil, err
	}
	if len(items) == 0 {
		return nil, errors.New("至少需要一个 SKU")
	}
	seen := map[string]bool{}
	skus := make([]model.ProductSKU, 0, len(items))
	for _, item := range items {
		specValues := item.SpecValues
		if len(specValues) == 0 && item.Spec != "" {
			specValues = map[string]string{"规格": item.Spec}
		}
		if len(dims) > 0 {
			for _, d := range dims {
				v, ok := specValues[d.Name]
				if !ok || strings.TrimSpace(v) == "" {
					return nil, fmt.Errorf("SKU 缺少规格「%s」", d.Name)
				}
				valid := false
				for _, allowed := range d.Values {
					if allowed == v {
						valid = true
						break
					}
				}
				if !valid {
					return nil, fmt.Errorf("SKU 规格值「%s=%s」不在定义范围内", d.Name, v)
				}
			}
		}
		key := specValuesKey(specValues, dims)
		if key == "" {
			return nil, errors.New("SKU 规格不能为空")
		}
		if seen[key] {
			return nil, fmt.Errorf("SKU 规格组合重复: %s", formatSpecSummary(specValues, dims))
		}
		seen[key] = true
		summary := item.Spec
		if summary == "" {
			summary = formatSpecSummary(specValues, dims)
		}
		if item.Price <= 0 {
			return nil, fmt.Errorf("SKU「%s」售价必须大于 0", summary)
		}
		if item.Stock < 0 {
			return nil, fmt.Errorf("SKU「%s」库存不能为负", summary)
		}
		skus = append(skus, model.ProductSKU{
			Spec:          summary,
			SpecValues:    marshalSpecValues(specValues),
			Price:         item.Price,
			OriginalPrice: item.OriginalPrice,
			Stock:         item.Stock,
			Image:         item.Image,
			Status:        1,
		})
	}
	return skus, nil
}
