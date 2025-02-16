package client

import (
	"ConjunctiveSSE/pkg/HDXT"
	pb "ConjunctiveSSE/pkg/HDXT/proto"
	"ConjunctiveSSE/pkg/utils"
	"context"
	"math"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type HDXTClient struct {
	hdxt   *HDXT.HDXT
	client pb.HDXTServiceClient
}

func NewHDXTClient(serverAddr string) (*HDXTClient, error) {
	conn, err := grpc.Dial(
		serverAddr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithDefaultCallOptions(
			grpc.MaxCallRecvMsgSize(100*1024*1024), // 100MB
			grpc.MaxCallSendMsgSize(100*1024*1024), // 100MB
		),
	)
	if err != nil {
		return nil, err
	}
	var hdxt HDXT.HDXT
	hdxt.Init("Crime_USENIX_REV", false)
	return &HDXTClient{
		hdxt:   &hdxt,
		client: pb.NewHDXTServiceClient(conn),
	}, nil
}

func (c *HDXTClient) GetHDXT() *HDXT.HDXT {
	return c.hdxt
}

func (c *HDXTClient) Setup(mitraCipherList map[string]string, auhmeCipherList map[string]string) error {
	// 发送密文到服务器
	stream, err := c.client.Setup(context.Background())
	if err != nil {
		return err
	}

	// 分批发送数据
    const batchSize = 1000
    count := 0
    batch := &pb.SetupRequest{
        MitraCiphers: make(map[string]string),
        AuhmeCiphers: make(map[string]string),
    }

    // 发送 MitraCiphers
    for k, v := range mitraCipherList {
        batch.MitraCiphers[k] = v
        count++
        
        if count >= batchSize {
            if err := stream.Send(batch); err != nil {
                return err
            }
            batch = &pb.SetupRequest{
                MitraCiphers: make(map[string]string),
                AuhmeCiphers: make(map[string]string),
            }
            count = 0
        }
    }

     // 发送 AuhmeCiphers
     for k, v := range auhmeCipherList {
        batch.AuhmeCiphers[k] = v
        count++
        
        if count >= batchSize {
            if err := stream.Send(batch); err != nil {
                return err
            }
            batch = &pb.SetupRequest{
                MitraCiphers: make(map[string]string),
                AuhmeCiphers: make(map[string]string),
            }
            count = 0
        }
    }

    // 发送最后一批数据并关闭流
    if count > 0 {
        if err := stream.Send(batch); err != nil {
            return err
        }
    }
    
    _, err = stream.CloseAndRecv()
	return err
}

func (c *HDXTClient) Update(id string, keywords []string, operation HDXT.Operation) error {
	// 本地生成密文
	_, tokList, err := c.hdxt.Encrypt(id, keywords, operation)
	if err != nil {
		return err
	}

	// 发送密文到服务器
	_, err = c.client.Update(context.Background(), &pb.UpdateRequest{
		UpdateTokens: convertToPbUTok(tokList),
	})
	return err
}

func (c *HDXTClient) SearchOneKeyword(keyword string) ([]string, error) {
	// 生成陷门
	tList, err := HDXT.MitraGenTrapdoor(c.hdxt, keyword)
	if err != nil {
		return nil, err
	}
	resp, err := c.client.SearchOneKeyword(context.Background(), &pb.SearchOneKeywordRequest{
		Trapdoors: tList,
	})
	if err != nil {
		return nil, err
	}
    ids, err := HDXT.MitraDecrypt(c.hdxt, keyword, resp.GetEncryptedIds())
    if err != nil {
        return nil, err
    }
    return ids, nil
}

func (c *HDXTClient) Search(keywords []string) ([]string, error) {
	// 单关键词搜索, mitra part
	// 选择查询频率最低的关键字
	counter, w1 := math.MaxInt64, keywords[0]
	for _, w := range keywords {
		num := c.hdxt.FileCnt[w]
		if num < counter {
			w1 = w
			counter = num
		}
	}
	w1Ids, err := c.SearchOneKeyword(w1)
	if err != nil {
		return nil, err
	}

	// auhme part
	// client search step 1
	q := utils.RemoveElement(keywords, w1)
	dkList, err := HDXT.AuhmeClientSearchStep1(c.hdxt, w1Ids, q)
	if err != nil {
		return nil, err
	}

	// 发送搜索请求
	resp, err := c.client.Search(context.Background(), &pb.SearchRequest{
		DkList: convertToPbDK(dkList),
	})
	if err != nil {
		return nil, err
	}

	// client search step 2
	sIdList := HDXT.AuhmeClientSearchStep2(w1Ids, convertToIntList(resp.PosList))
	return sIdList, nil
}
