package pdf

import (
	"bytes"
	"fmt"
	"image"
	"image/png"

	"github.com/epalmerini/mtg-proxy/internal/card"
	"github.com/go-pdf/fpdf"
)

const (
	pageW=210.0; pageH=297.0; cardW=63.0; cardH=88.0; cols=3; rows=3; cardsPerPage=cols*rows
	marginX=(pageW-cols*cardW)/2; marginY=(pageH-rows*cardH)/2; padding=2.0; fontName="dejavu"
)
type Renderer struct{compress bool}
type Option func(*Renderer)
func WithCompression(on bool)Option{return func(r *Renderer){r.compress=on}}
func NewRenderer(opts ...Option)*Renderer{r:=&Renderer{compress:true};for _,o:=range opts{o(r)};return r}
func(r *Renderer)Render(cards []card.DeckCard,outputPath string)error{p:=fpdf.New("P","mm","A4","");p.SetCompression(r.compress);p.SetAutoPageBreak(false,0);p.AddUTF8FontFromBytes(fontName,"",dejaVuRegular);p.AddUTF8FontFromBytes(fontName,"B",dejaVuBold);expanded:=expandDeck(cards);if len(expanded)==0{p.AddPage()};for i,ec:=range expanded{if i%cardsPerPage==0{p.AddPage();n:=len(expanded)-i;if n>cardsPerPage{n=cardsPerPage};drawGrid(p,n)};pos:=i%cardsPerPage;renderCard(p,ec,marginX+float64(pos%cols)*cardW,marginY+float64(pos/cols)*cardH)};return p.OutputFileAndClose(outputPath)}
type expandedCard struct{card.Card;artImages []image.Image}
func expandDeck(cards []card.DeckCard)[]expandedCard{var out []expandedCard;for _,dc:=range cards{images:=dc.ArtImages;if len(images)==0&&dc.ArtImage!=nil{images=[]image.Image{dc.ArtImage}};for i:=0;i<int(dc.Quantity);i++{out=append(out,expandedCard{Card:dc.Card,artImages:images})}};return out}
func drawGrid(p *fpdf.Fpdf,cardCount int){p.SetDrawColor(0,0,0);p.SetLineWidth(0.2);p.SetDashPattern([]float64{0.3,1.5},0);usedRows:=(cardCount+cols-1)/cols;for r:=0;r<=usedRows;r++{y:=marginY+float64(r)*cardH;n:=cols;if r==usedRows{n=cardCount-(usedRows-1)*cols};p.Line(marginX,y,marginX+float64(n)*cardW,y)};for r:=0;r<usedRows;r++{n:=cols;if r==usedRows-1{n=cardCount-r*cols};for c:=0;c<=n;c++{x:=marginX+float64(c)*cardW;y:=marginY+float64(r)*cardH;p.Line(x,y,x,y+cardH)}};p.SetDashPattern([]float64{},0)}
func renderCard(p *fpdf.Fpdf,ec expandedCard,x,y float64){if ec.IsMultiFaced(){halfH:=(cardH-padding)/2;renderFace(p,ec.Faces[0],x,y,halfH,imageAt(ec.artImages,0));sepY:=y+halfH+padding/2;p.SetDrawColor(0,0,0);p.SetLineWidth(0.2);p.SetDashPattern([]float64{0.3,1.5},0);p.Line(x+padding,sepY,x+cardW-padding,sepY);p.SetDashPattern([]float64{},0);faceY:=sepY+padding/2;cx:=x+cardW/2;cy:=faceY+halfH/2;p.TransformBegin();p.TransformRotate(180,cx,cy);renderFace(p,ec.Faces[1],x,faceY,halfH,imageAt(ec.artImages,1));p.TransformEnd()}else{renderFace(p,ec.Front(),x,y,cardH,imageAt(ec.artImages,0))}}
func imageAt(images []image.Image,i int)image.Image{if i<0||i>=len(images){return nil};return images[i]}
var artCounter int
func renderFace(p *fpdf.Fpdf,f card.CardFace,x,y,height float64,artImage image.Image){innerX:=x+padding;innerW:=cardW-2*padding;cursorY:=y+padding;p.SetFont(fontName,"B",7);p.SetXY(innerX,cursorY);mana:=f.ManaCost.String();nameW:=innerW;if mana!=""{costW:=p.GetStringWidth(mana)+1;nameW=innerW-costW;p.CellFormat(nameW,4,string(f.Name),"",0,"L",false,0,"");p.CellFormat(costW,4,mana,"",0,"R",false,0,"")}else{p.CellFormat(nameW,4,string(f.Name),"",0,"L",false,0,"")};cursorY+=5;p.SetDrawColor(0,0,0);p.SetLineWidth(0.1);p.SetDashPattern([]float64{0.3,1.5},0);p.Line(innerX,cursorY,x+cardW-padding,cursorY);p.SetDashPattern([]float64{},0);cursorY+=1;if artImage!=nil{artH:=height*0.30;if artH>30{artH=30};if cursorY+artH>y+height-padding{artH=y+height-padding-cursorY};if artH>0{registerAndPlaceImage(p,artImage,innerX,cursorY,innerW,artH);cursorY+=artH+1;p.SetDashPattern([]float64{0.3,1.5},0);p.Line(innerX,cursorY-1,x+cardW-padding,cursorY-1);p.SetDashPattern([]float64{},0)}};p.SetFont(fontName,"",6);p.SetXY(innerX,cursorY);p.CellFormat(innerW,3.5,string(f.TypeLine),"",0,"L",false,0,"");cursorY+=4.5;p.SetDashPattern([]float64{0.3,1.5},0);p.Line(innerX,cursorY,x+cardW-padding,cursorY);p.SetDashPattern([]float64{},0);cursorY+=1;footerH:=0.0;if f.Stats!=nil||f.Loyalty!=nil{footerH=5};oracleMaxY:=y+height-padding-footerH;if oracleMaxY>cursorY{p.SetFont(fontName,"",5.5);p.ClipRect(innerX,cursorY,innerW,oracleMaxY-cursorY,false);p.SetXY(innerX,cursorY);p.MultiCell(innerW,3,string(f.OracleText),"","L",false);p.ClipEnd()};if f.Stats!=nil{footer:=fmt.Sprintf("%s/%s",f.Stats.Power,f.Stats.Toughness);p.SetFont(fontName,"B",7);p.SetXY(innerX,y+height-padding-4);p.CellFormat(innerW,4,footer,"",0,"R",false,0,"")}else if f.Loyalty!=nil{footer:=fmt.Sprintf("Loyalty: %s",*f.Loyalty);p.SetFont(fontName,"B",7);p.SetXY(innerX,y+height-padding-4);p.CellFormat(innerW,4,footer,"",0,"R",false,0,"")}}
func registerAndPlaceImage(p *fpdf.Fpdf,img image.Image,x,y,w,h float64){var buf bytes.Buffer;_ = png.Encode(&buf,img);name:=fmt.Sprintf("art_%d",artCounter);artCounter++;p.RegisterImageOptionsReader(name,fpdf.ImageOptions{ImageType:"PNG"},&buf);p.ImageOptions(name,x,y,w,h,false,fpdf.ImageOptions{ImageType:"PNG"},0,"")}
