(defpurefun ((vanishes! :𝔽@loob) x) x)

(defcolumns (A :i16@loob) (B :i16))
(defconstraint c1 ()
  (let ((C A) (D B))
    (if C
        (vanishes! D))))
