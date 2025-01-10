(defpurefun ((vanishes! :@loob) x) x)

(defcolumns (A :@loob) B)
(defconstraint c1 ()
  (let ((B B))
    (if A
        (vanishes! B))))
