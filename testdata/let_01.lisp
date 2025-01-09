(defpurefun ((vanishes! :@loob) x) x)

(defcolumns (A :@loob) B)
(defconstraint c1 ()
  (let ((C B))
    (if A
        (vanishes! C))))
