(defpurefun ((vanishes! :𝔽@loob) x) x)

(defcolumns (STAMP :i12))
;; STAMP[-1] == 0
(defconstraint c1 (:domain {-1}) (vanishes! STAMP))
