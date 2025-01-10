(defpurefun ((vanishes! :@loob) x) x)

(defcolumns (A :@loob) B)
(defconstraint c1 ()
  (let ((C A) (D B))
    (if C
        (vanishes! D))))
