(defconst (CHAIN :extern) 1)

(defconst
  LIMIT_0 1000
  LIMIT_1 1100)

(defpurefun (LIMIT) (+
           ;; CHAIN=0
           (* (- 1 CHAIN) LIMIT_0)
           ;; CHAIN=1
           (* CHAIN LIMIT_1)))

(defcolumns ST (X :@loob))

(defconstraint c1 (:guard ST) (- X (LIMIT)))
