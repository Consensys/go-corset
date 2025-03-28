;; An example of a simple memory implemented using permutation sorts.
;; Each row in the trace represents a read (RW=0) or write (RW=1) to
;; memory.  For a read, the VAL column identifies the value read at
;; ADDR (i.e. the read address).  For a write, the VAL column holds
;; the value being written at ADDR (i.e. the write address).
;;
;; The constriants should enforce that we cannot construct values "out
;; of thin air".  That is, every read for a given address matches the
;; last write (or 0 if there was no previous write).  More
;; specifically, the most recent PC value where that addres was
;; written.

;; Program Counter (always increases by one)
(defcolumns (PC :i16@prove))
;; Read/Write flag (0=READ, 1=WRITE)
(defcolumns (RW :i1@prove))
;; Address being Read/Written
(defcolumns (ADDR :i32@prove))
;; Value being Read/Written
(defcolumns (VAL :i8@prove))

;; Permutation
(defpermutation (ADDR' PC' RW' VAL') ((+ ADDR) (+ PC) (+ RW) (+ VAL)))

;; PC[0]=0
(defconstraint heartbeat_1 (:domain {0}) (eq! PC 0))

;; PC[k]=0 || PC[k]=PC[k-1]+1
(defconstraint heartbeat_2 ()
  (or!
   (eq! PC 0)
   (eq! PC (+ 1 (prev PC)))))

;; PC[k]=0 ==> PC[k-1]=0
(defconstraint heartbeat_3 ()
  (if (== 0 PC)
      (eq! (prev PC) 0)))

;; PC[k]=0 ==> (RW[k]=0 && ADDR[k]=0 && VAL[k]=0)
(defconstraint heartbeat_4 ()
  (if (== 0 PC)
      (begin
       (eq! RW 0)
       (eq! ADDR 0)
       (eq! VAL 0))))

;; ADDR'[k] != ADDR'[k-1] ==> (RW'[k]=1 || VAL'[k]=0)
(defconstraint first_read_1 ()
  (if-not-eq ADDR' (prev ADDR')
      (or!
       (eq! RW' 1)
       (eq! VAL' 0))))

;; (RW'[0]=1 || VAL'[0]=0)
(defconstraint first_read_2 (:domain {0})
  (or!
   (eq! RW' 1)
   (eq! VAL' 0)))

;; ADDR'[k] == ADDR'[k-1] ==> (RW=1 || VAL'[k]=VAL'[k-1])
(defconstraint next_read ()
  (if
   (eq! ADDR' (prev ADDR'))
   (or!
    (eq! RW' 1)
    (eq! VAL' (prev VAL')))))
