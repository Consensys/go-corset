(defpurefun ((vanishes! :@loob) x) x)

(defcolumns ST A B)
(defconstraint c1 () (vanishes! (* ST (- 1 (+ (~ A) (~ B))))))
