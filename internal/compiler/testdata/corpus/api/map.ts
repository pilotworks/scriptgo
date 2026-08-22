// ScriptGo Corpus: Map Standard Builtin APIs
// Consolidated test suite with inline assertions.

// @api: map.clear
// @expect: 0
const m_map_clear_0 = new Map<string, number>(); m_map_clear_0.set("a", 1); m_map_clear_0.clear(); console.log(m_map_clear_0.size);

// @api: map.delete
// @expect: true
// @expect: false
const m_map_delete_1 = new Map<string, number>(); m_map_delete_1.set("a", 1); console.log(m_map_delete_1.delete("a")); console.log(m_map_delete_1.has("a"));

// @api: map.forEach
// @expect: x=5
const m_map_forEach_2 = new Map<string, number>();
m_map_forEach_2.set("x", 5);
m_map_forEach_2.forEach((val: number, key: string) => {
    console.log(key + "=" + val);
});

// @api: map.get
// @expect: 100
const m_map_get_3 = new Map<string, number>(); m_map_get_3.set("a", 100); console.log(m_map_get_3.get("a"));

// @api: map.has
// @expect: true
// @expect: false
const m_map_has_4 = new Map<string, number>(); m_map_has_4.set("a", 1); console.log(m_map_has_4.has("a")); console.log(m_map_has_4.has("b"));

// @api: map.set
// @expect: 1
const m_map_set_5 = new Map<string, number>(); m_map_set_5.set("a", 1); console.log(m_map_set_5.size);
