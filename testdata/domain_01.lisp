(defpurefun ((vanishes! :𝔽@loob) x) x)

(defcolumns (STAMP :i10))
;; STAMP[0] == 0
(defconstraint c1 (:domain {0}) (vanishes! STAMP))
