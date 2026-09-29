package catalog_test
import("html-ui/internal/catalog";"testing")
func TestDisclosureContentPart(t *testing.T){
 for _,name:=range []string{"accordion","collapsible"}{
  r,_:=catalog.Find(name)
  part,ok:=r.UI.Parts["content"]
  if !ok || part.Node!="content" || part.State["expanded"].Source.Node!="root"{t.Fatalf("%s needs owned content bound to root.open",name)}
  found:=false
  for _,c:=range r.Children{if c.Parent=="content"&&c.Slot=="content"{found=true}}
  if !found{t.Fatalf("%s content slot must belong to content wrapper",name)}
 }
}
