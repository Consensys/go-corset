(defpurefun ((vanishes! :@loob) x) x)

(defcolumns A)
(defpurefun (id x) x)
(defunalias ID id)
(defconstraint test () (vanishes! (ID A)))
