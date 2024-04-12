package message

func _getRawMessage() map[string][]string {
	raw_message := make(map[string][]string)
	raw_message["basic"] = _getBasicRawMessage()
	raw_message["resource"] = _getResourceRawMessage()
	raw_message["decorate"] = _getDecorateRawMessage()
	raw_message["layout"] = _getLayoutRawMessage()
	raw_message["meta"] = _getMetaRawMessage()
	raw_message["interact"] = _getInteractRawMessage()
	raw_message["extend"] = _getExtendRawMessage()
	return raw_message
}

func _getBasicRawMessage() []string {
	return []string{
		// Single message
		`I can eat glass and it doesn't hurt me.`,
		`<at id="test" name="test" role="test" type="test"/>`,
		`<sharp id="test" name="test"/>`,
		`<a href="https://example.com"/>`,
	}
}

func _getResourceRawMessage() []string {
	return []string{
		`<img src="https://example.com" title="exampleImage" cache timeout="1145141919" width="200" height="200"/>`,
		`<audio src="https://example.com" title="exampleAudio" cache timeout="1145141919" duration="114514" poster="https://example.com"/>`,
		`<video src="https://example.com" title="exampleVideo" cache timeout="1145141919" width="200" height="200" duration="114514" poster="https://example.com"/>`,
		`<file src="https://example.com" title="exampleFile" cache timeout="1145141919" poster="https://example.com"/>`,
	}
}

func _getDecorateRawMessage() []string {
	return []string{
		`<b>b<i>italic</i>o<u>under<s>deleteline</s>line</u>l<spl>sp<code>code</code>oi<sup>s<sub>sub</sub>up</sup>ler</spl>d</b>`,
		`<code>for key, value range attributes {
			fmt.Println(key, value)
		}</code>`,
	}
}

func _getLayoutRawMessage() []string {
	return []string{
		`<br/>`,
		`<p>test</p>`,
		`<message id="test" forward><p>test</p>test<br/><p>test<br/></p></message>`,
		`test<message/>test`,
	}
}

func _getMetaRawMessage() []string {
	return []string{
		`<quote><author id="test" name="test">test</author>test</quote>`,
		`<author id="test" name="test" avatar="https://example.com"/>`,
	}
}

func _getInteractRawMessage() []string {
	return []string{
		`<button id="test" type="test" href="https://example.com" text="test" theme="test"/>`,
	}
}

func _getExtendRawMessage() []string {
	return []string{
		`<audio src="https://example.com" test:test="test"/>`,
		`<video src="https://example.com" test:test/>`,
		`<img src="https://example.com">test</img>`,
		`<file src="https://example.com" test2:test2>test</file>`,
		`<test:test test="test">test</test:test>`,
	}
}
