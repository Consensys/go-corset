(defpurefun ((vanishes! :@loob) x) x)

(defcolumns A)
(defpurefun (_id x) x)
(defunalias ID _id)
(defconstraint test () (vanishes! (ID A)))
