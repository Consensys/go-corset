;; simple definition of a Read-Only Memory (ROM)
(defcolumns (enable :u1@prove) (address :u16@prove) (data :u8@prove))
;; ensure enable goes high only once, and that address is 0
(defconstraint enable ()
  (if (!= enable (prev enable))
     (∧
      ;; enable has rising edge
      (== 1 enable)
      ;; addresses start at zero
      (== 0 address))))
;; ensure address always 0 in padding region
(defconstraint padding ()
  (if (== 0 enable)
      (== address 0)))
;; ensure address always increments
(defconstraint clk ()
  (if (== 1 enable)
      (== (+ 1 address) (next address))))
