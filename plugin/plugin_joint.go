// Package plugin
// @Author: yinwei
// @File: plugin_joint
// @Version: 1.0.0
// @Date: 2024/10/30 10:11

package plugin

import (
	"bytes"
	"fmt"
	validator "github.com/coderyw/go-proto-validators"
	"github.com/coderyw/protobuf/gogoproto"
	"github.com/coderyw/protobuf/protoc-gen-gogo/descriptor"
	"github.com/coderyw/protobuf/protoc-gen-gogo/generator"
	"os"
	"strings"
)

func writeBuffer(bf *bytes.Buffer, str ...interface{}) {
	for _, v := range str {
		bf.WriteString(fmt.Sprint(v))
	}
}

func (p *plugin) generateJint(myField *descriptor.FieldDescriptorProto, fieldValidator *validator.FieldValidator, importPath generator.GoImportPath, file *generator.FileDescriptor, message *generator.Descriptor) {
	ccTypeName := generator.CamelCaseSlice(message.TypeName())
	buffers := &bytes.Buffer{}
	buffers.WriteString(`if true`)
	var errorStr = &strings.Builder{}
	writeErr := func(b *strings.Builder, str ...string) {
		if b.Len() > 0 {
			b.WriteString(" and ")
		}
		b.WriteString(strings.Join(str, " and "))
	}
	for _, v := range fieldValidator.IfField {
		if v.Validator == nil {
			continue
		}
		for _, field := range message.Field {

			// 找到对应的变量
			if field.Name == nil || *field.Name != *v.Field {
				continue
			}
			//fieldValidator := getFieldValidatorIfAny(field)
			fieldName := p.GetOneOfFieldName(message, field)
			variableName := "this." + fieldName
			repeated := field.IsRepeated()
			// Golang's proto3 has no concept of unset primitive fields
			nullable := (gogoproto.IsNullable(field) || !gogoproto.ImportsGoGoProto(file.FileDescriptorProto)) && field.IsMessage() && !(p.useGogoImport && gogoproto.IsEmbed(field))
			if p.fieldIsProto3Map(file, message, field) {
				p.P(`// Validation of proto3 map<> fields is unsupported.`)
				continue
			}

			if repeated {
				if v.Validator.Required != nil && *v.Validator.Required {
					writeBuffer(buffers, ` && len(`, variableName, `)!=0`)
					writeErr(errorStr, fmt.Sprintf(`len(%v) != 0`, fieldName))
				}
				if v.Validator.RepeatedCountMin != nil {
					writeBuffer(buffers, `&& len(`, variableName, `) >= `, *v.Validator.RepeatedCountMin)
					writeErr(errorStr, fmt.Sprintf(`len(%v) >= 0`, fieldName))
				}
				if v.Validator.RepeatedCountMax != nil {
					writeBuffer(buffers, `&& len(`, variableName, `) <= `, *v.Validator.RepeatedCountMax)
					writeErr(errorStr, fmt.Sprintf(`len(%v) <= 0`, fieldName))
				}
			}
			if field.IsString() {
				if v.Validator.StringLengthGt != nil {
					writeBuffer(buffers, ` && (int64(`, p.utf8Pkg.Use(), `.RuneCountInString(`, variableName, `))>`, *v.Validator.StringLengthGt, `)`)
					writeErr(errorStr, fmt.Sprintf(`%v's length > %v`, fieldName, v.Validator.GetStringLengthGt()))
				}
				if v.Validator.StringLengthLt != nil {
					writeBuffer(buffers, ` && (int64(`, p.utf8Pkg.Use(), `.RuneCountInString(`, variableName, `))<`, *v.Validator.StringLengthLt, `)`)
					writeErr(errorStr, fmt.Sprintf(`%v's length < %v`, fieldName, v.Validator.GetStringLengthLt()))
				}
				if v.Validator.StringLengthEq != nil {
					writeBuffer(buffers, ` && (int64(`, p.utf8Pkg.Use(), `.RuneCountInString(`, variableName, `))==`, *v.Validator.StringLengthEq)
					writeErr(errorStr, fmt.Sprintf(`%v's length == %v`, fieldName, v.Validator.GetStringLengthEq()))
				}
				if v.Validator.Regex != nil {
					writeBuffer(buffers, ` && `, p.regexName(ccTypeName, fmt.Sprintf("%v_%v", fieldName, *v.Field)), `.MatchString(`, variableName, `)`)
					writeErr(errorStr, fmt.Sprintf(`%v match with %v`, fieldName, v.Validator.GetStringLengthGt()))
				}
				if v.Validator.StringNotEmpty != nil && v.Validator.GetStringNotEmpty() {
					writeBuffer(buffers, ` && `, variableName, ` != ""`)
					writeErr(errorStr, fmt.Sprintf(`%v is't empty`, fieldName))
				}
				if v.Validator.Required != nil && *v.Validator.Required {
					writeBuffer(buffers, ` && `, variableName, ` != ""`)
					writeErr(errorStr, fmt.Sprintf(`%v is't empty`, fieldName))
				}
				if v.Validator.LengthGt != nil {
					writeBuffer(buffers, ` && len(`, variableName, `) > `, *v.Validator.LengthGt)
					writeErr(errorStr, fmt.Sprintf(`%v's length > %v`, fieldName, v.Validator.GetLengthGt()))
				}
				if v.Validator.LengthLt != nil {
					writeBuffer(buffers, ` && len(`, variableName, `) < `, *v.Validator.LengthLt)
					writeErr(errorStr, fmt.Sprintf(`%v's length < %v`, fieldName, v.Validator.GetLengthLt()))
				}
				if v.Validator.LengthEq != nil {
					writeBuffer(buffers, ` && len(`, variableName, `) == `, *v.Validator.LengthEq)
					writeErr(errorStr, fmt.Sprintf(`%v's length = %v`, fieldName, v.Validator.GetLengthEq()))
				}
				if v.Validator.StringEq != nil {
					writeBuffer(buffers, ` && `, variableName, ` == "`, *v.Validator.StringEq, `"`)
					writeErr(errorStr, fmt.Sprintf(`%v's value = '%v'`, fieldName, v.Validator.GetLengthGt()))
				}

			} else if p.isSupportedInt(field) {
				if v.Validator.Required != nil && *v.Validator.Required {
					writeBuffer(buffers, ` && `, variableName, `==0`)
					writeErr(errorStr, fmt.Sprintf(`%v's value != 0`, fieldName))
				}
				if v.Validator.IntGt != nil {
					writeBuffer(buffers, ` && `, variableName, `>`, *v.Validator.IntGt)
					writeErr(errorStr, fmt.Sprintf(`%v's value > %v`, fieldName, v.Validator.GetIntGt()))
				}
				if v.Validator.IntLt != nil {
					writeBuffer(buffers, ` && `, variableName, `<`, *v.Validator.IntLt)
					writeErr(errorStr, fmt.Sprintf(`%v's value < %v`, fieldName, v.Validator.GetIntLt()))
				}
				if v.Validator.IntEq != nil {
					writeBuffer(buffers, ` && `, variableName, `==`, *v.Validator.IntEq)
					writeErr(errorStr, fmt.Sprintf(`%v's value = %v`, fieldName, v.Validator.GetIntEq()))
				}
			} else if field.IsEnum() {
				if v.Validator.Required != nil && *v.Validator.Required {
					writeBuffer(buffers, ` && `, variableName, `==0`)
					writeErr(errorStr, fmt.Sprintf(`%v's value != 0`, fieldName))
				}
				if v.Validator.IntGt != nil {
					writeBuffer(buffers, ` && `, variableName, `>`, *v.Validator.IntGt)
					writeErr(errorStr, fmt.Sprintf(`%v's value > %v`, fieldName, v.Validator.GetIntGt()))
				}
				if v.Validator.IntLt != nil {
					writeBuffer(buffers, ` && `, variableName, `<`, *v.Validator.IntLt)
					writeErr(errorStr, fmt.Sprintf(`%v's value < %v`, fieldName, v.Validator.GetIntLt()))
				}
				if v.Validator.IntEq != nil {
					writeBuffer(buffers, ` && `, variableName, `==`, *v.Validator.IntEq)
					writeErr(errorStr, fmt.Sprintf(`%v's value = %v`, fieldName, v.Validator.GetIntEq()))
				}
			} else if p.isSupportedFloat(field) {
				upperIsStrict := true
				lowerIsStrict := true
				if v.Validator.Required != nil && *v.Validator.Required {
					writeBuffer(buffers, ` && `, variableName, `==0`)
					writeErr(errorStr, fmt.Sprintf(`%v's value != 0`, fieldName))
				}
				if v.Validator.FloatEpsilon != nil && v.Validator.FloatLt == nil && v.Validator.FloatGt == nil {
					fmt.Fprintf(os.Stderr, "WARNING: field %v.%v has no 'float_lt' or 'float_gt' field so setting 'float_epsilon' has no effect.", ccTypeName, fieldName)
				}
				if v.Validator.FloatLt != nil && v.Validator.FloatLte != nil {

					fmt.Fprintf(os.Stderr, "WARNING: field %v.%v has both 'float_lt' and 'float_lte' constraints, only the strictest will be used.", ccTypeName, fieldName)
					strictLimit := v.Validator.GetFloatLt()
					if v.Validator.FloatEpsilon != nil {
						strictLimit += v.Validator.GetFloatEpsilon()
					}

					if v.Validator.GetFloatLte() < strictLimit {
						upperIsStrict = false
					}
				} else if v.Validator.FloatLte != nil {
					upperIsStrict = false
				}
				if v.Validator.FloatGt != nil && v.Validator.FloatGte != nil {
					fmt.Fprintf(os.Stderr, "WARNING: field %v.%v has both 'float_gt' and 'float_gte' constraints, only the strictest will be used.", ccTypeName, fieldName)
					strictLimit := v.Validator.GetFloatGt()
					if v.Validator.FloatEpsilon != nil {
						strictLimit -= v.Validator.GetFloatEpsilon()
					}

					if v.Validator.GetFloatGte() > strictLimit {
						lowerIsStrict = false
					}
				} else if v.Validator.FloatGte != nil {
					lowerIsStrict = false
				}

				compareStr := ""
				if v.Validator.FloatGt != nil || v.Validator.FloatGte != nil {
					compareStr = fmt.Sprint(` && `, variableName)
					if lowerIsStrict {
						if v.Validator.FloatEpsilon != nil {
							compareStr += fmt.Sprint(` + `, v.Validator.GetFloatEpsilon())
						}
						compareStr += fmt.Sprint(` > `, v.Validator.GetFloatGt())
						writeErr(errorStr, fmt.Sprintf(`%v's value > %v`, variableName, v.Validator.GetFloatGt()))
					} else {
						compareStr += fmt.Sprint(` >= `, v.Validator.GetFloatGte())
						writeErr(errorStr, fmt.Sprintf(`%v's value >= %v`, variableName, v.Validator.GetFloatGte()))
					}
					writeBuffer(buffers, compareStr)
				}

				if v.Validator.FloatLt != nil || v.Validator.FloatLte != nil {
					compareStr = fmt.Sprint(` && `, variableName)
					if upperIsStrict {
						if v.Validator.FloatEpsilon != nil {
							compareStr += fmt.Sprint(` - `, v.Validator.GetFloatEpsilon())
						}
						compareStr += fmt.Sprint(` < `, v.Validator.GetFloatLt())
						writeErr(errorStr, fmt.Sprintf(`%v's value < %v`, fieldName, v.Validator.GetFloatLt()))
					} else {
						compareStr += fmt.Sprint(` <= `, v.Validator.GetFloatLte())
						writeErr(errorStr, fmt.Sprintf(`%v's value <= %v`, fieldName, v.Validator.GetFloatLte()))
					}
				}

			} else if field.IsBytes() {
				if v.Validator.Required != nil && *v.Validator.Required {
					writeBuffer(buffers, ` && len(`, variableName, `)!=0`)
					writeErr(errorStr, fmt.Sprintf(`%v's value is not empty`, fieldName))
				}

				if v.Validator.LengthGt != nil {
					writeBuffer(buffers, ` && ( len(`, variableName, `) > `, v.Validator.LengthGt)
					writeErr(errorStr, fmt.Sprintf(`%v's length > %v`, fieldName, v.Validator.GetLengthGt()))
				}

				if v.Validator.LengthLt != nil {
					writeBuffer(buffers, ` && ( len(`, variableName, `) < `, v.Validator.LengthLt)
					writeErr(errorStr, fmt.Sprintf(`%v's length < %v`, fieldName, v.Validator.GetLengthLt()))
				}

				if v.Validator.LengthEq != nil {
					writeBuffer(buffers, ` && ( len(`, variableName, `) == `, v.Validator.LengthEq)
					writeErr(errorStr, fmt.Sprintf(`%v's length = %v`, fieldName, v.Validator.GetLengthEq()))
				}
			} else if field.IsMessage() {
				if (v.Validator.MsgExists != nil && *v.Validator.MsgExists) || (v.Validator.Required != nil && *v.Validator.Required) {
					if nullable && !repeated {
						writeBuffer(buffers, ` && nil != `, variableName)
						writeErr(errorStr, fmt.Sprintf(`%v is not nil`, variableName))
					} else if repeated {
						fmt.Fprintf(os.Stderr, "WARNING: field %v.%v is repeated, validator.msg_exists has no effect\n", ccTypeName, fieldName)
					} else if !nullable {
						fmt.Fprintf(os.Stderr, "WARNING: field %v.%v is a nullable=false, validator.msg_exists has no effect\n", ccTypeName, fieldName)
					}
				}

			}

		}
	}
	writeBuffer(buffers, "{")
	p.P(buffers.String())
	p.In()
	var es string
	if errorStr.Len() > 0 {
		es = "When " + errorStr.String() + ", "
	}
	p.generateOne(file, message, importPath, myField, fieldValidator, ccTypeName, es)
	p.Out()
	p.P(`}`)
}
